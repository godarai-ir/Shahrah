#!/usr/bin/env node

import { lstat, readdir, readFile, rm } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const repositoryRoot = path.resolve(scriptDirectory, "..");
const mode = process.argv[2] ?? "stale";
const dryRun = process.argv.includes("--dry-run");
const force = process.argv.includes("--force");
const supportedModes = new Set([
	"stale",
	"cache",
	"next",
	"temp",
	"all",
	"prune",
]);

if (!supportedModes.has(mode)) {
	fail(
		`Unsupported mode: ${mode}. Use stale, cache, next, temp, all, or prune.`,
	);
}

await verifyRepositoryRoot();

if (mode === "prune") {
	await pruneTurboCache();
	process.exit(0);
}

const selection = await collectTargets(mode);
const targets = selection.targets;
let reclaimedBytes = 0;
let skippedLockedTarget = false;

for (const target of targets) {
	assertSafeTarget(target);
	const size = await getPathSize(target);
	if (size === null) {
		continue;
	}

	console.log(
		`${dryRun ? "Would remove" : "Removing"} ${relativePath(target)} (${formatBytes(size)})${selection.reasons.get(target) ? `: ${selection.reasons.get(target)}` : ""}`,
	);
	if (!dryRun) {
		try {
			await rm(target, {
				force: true,
				recursive: true,
				maxRetries: 3,
				retryDelay: 250,
			});
		} catch (error) {
			if (!["EBUSY", "ENOTEMPTY", "EPERM"].includes(error.code)) {
				throw error;
			}

			const remainingBytes = (await getPathSize(target)) ?? 0;
			reclaimedBytes += Math.max(0, size - remainingBytes);
			skippedLockedTarget = true;
			console.warn(
				`Skipped locked content in ${relativePath(target)} (${formatBytes(remainingBytes)} remaining): ${error.message}`,
			);
			continue;
		}
	}
	reclaimedBytes += size;
}

console.log(
	`${dryRun ? "Potential reclaim" : "Reclaimed"}: ${formatBytes(reclaimedBytes)}`,
);
let remainingManagedBytes = 0;
for (const target of selection.managedTargets) {
	remainingManagedBytes += (await getPathSize(target)) ?? 0;
}
const reportedRemainingBytes = dryRun
	? Math.max(0, remainingManagedBytes - reclaimedBytes)
	: remainingManagedBytes;
console.log(
	`${dryRun ? "Projected remaining managed data" : "Remaining managed data"}: ${formatBytes(reportedRemainingBytes)}`,
);
if (skippedLockedTarget) {
	process.exitCode = 1;
}

async function collectTargets(selectedMode) {
	const reasons = new Map();
	const staleTargets = await findNextStaleDirectories();
	if (selectedMode === "stale") {
		return createSelection(staleTargets, staleTargets, reasons);
	}

	const nextCacheSelection = await findPrunableNextCaches(force);
	for (const [target, reason] of nextCacheSelection.reasons) {
		reasons.set(target, reason);
	}
	if (selectedMode === "next") {
		return createSelection(
			nextCacheSelection.targets,
			nextCacheSelection.managedTargets,
			reasons,
		);
	}

	const temporaryTargets = [
		resolveFromRoot(".codex", "tmp"),
		resolveFromRoot(".golangci-lint-report.json"),
		resolveFromRoot("GodarAI-Web-Assets.zip"),
		resolveFromRoot("bin"),
		resolveFromRoot("services", "control-plane", "bin"),
		resolveFromRoot("services", "control-plane", "tenant-purge.exe"),
	];
	if (selectedMode === "temp") {
		return createSelection(temporaryTargets, temporaryTargets, reasons);
	}

	const baseCacheTargets = [
		...staleTargets,
		resolveFromRoot(".turbo"),
		resolveFromRoot(".gocache"),
		resolveFromRoot(".pnpm-store"),
		resolveFromRoot(".pnpm-store-codex"),
		resolveFromRoot(".ruff_cache"),
		resolveFromRoot(".cache"),
		resolveFromRoot(".tmp"),
		resolveFromRoot("services", "control-plane", ".tmp"),
	];
	const cacheTargets = [
		...baseCacheTargets,
		...nextCacheSelection.targets,
	];

	if (selectedMode === "cache") {
		return createSelection(
			cacheTargets,
			[...cacheTargets, ...nextCacheSelection.managedTargets],
			reasons,
		);
	}

	const generatedTargets = [
		resolveFromRoot("coverage"),
		resolveFromRoot("bin"),
	];
	const appsDirectory = resolveFromRoot("apps");
	for (const entry of await readDirectorySafely(appsDirectory)) {
		if (!entry.isDirectory()) {
			continue;
		}

		const appDirectory = path.join(appsDirectory, entry.name);
		for (const generatedName of [
			".astro",
			".docusaurus",
			".next",
			".open-next",
			"build",
			"coverage",
			"dist",
		]) {
			generatedTargets.push(path.join(appDirectory, generatedName));
		}
	}

	return createSelection(
		[...baseCacheTargets, ...temporaryTargets, ...generatedTargets],
		[
			...baseCacheTargets,
			...temporaryTargets,
			...generatedTargets,
		],
		reasons,
	);
}

async function findPrunableNextCaches(forceRemoval) {
	const targets = [];
	const managedTargets = [];
	const reasons = new Map();
	const maximumGiB = readPositiveNumber(
		"SHAHRAH_NEXT_CACHE_MAX_GIB",
		"2",
	);
	const maximumAgeDays = readPositiveNumber(
		"SHAHRAH_NEXT_CACHE_MAX_AGE_DAYS",
		"14",
	);
	const maximumBytes = maximumGiB * 1024 ** 3;
	const oldestAllowedTime = Date.now() - maximumAgeDays * 24 * 60 * 60 * 1000;
	const appsDirectory = resolveFromRoot("apps");

	for (const appEntry of await readDirectorySafely(appsDirectory)) {
		if (!appEntry.isDirectory()) {
			continue;
		}
		for (const segments of [
			[".next", "cache"],
			[".next", "dev", "cache"],
		]) {
			const cacheDirectory = path.join(
				appsDirectory,
				appEntry.name,
				...segments,
			);
			const stats = await getPathStats(cacheDirectory);
			if (stats === null) {
				continue;
			}
			managedTargets.push(cacheDirectory);
			const tooLarge = stats.size > maximumBytes;
			const tooOld = stats.latestModifiedAt < oldestAllowedTime;
			if (!forceRemoval && !tooLarge && !tooOld) {
				continue;
			}
			targets.push(cacheDirectory);
			const reason = forceRemoval
				? "forced"
				: tooLarge
					? `exceeds ${maximumGiB} GiB limit`
					: `older than ${maximumAgeDays} days`;
			reasons.set(path.resolve(cacheDirectory), reason);
		}
	}

	return {
		targets: uniquePaths(targets),
		managedTargets: uniquePaths(managedTargets),
		reasons,
	};
}

async function findNextStaleDirectories() {
	const targets = [];
	const appsDirectory = resolveFromRoot("apps");

	for (const appEntry of await readDirectorySafely(appsDirectory)) {
		if (!appEntry.isDirectory()) {
			continue;
		}

		const nextDirectory = path.join(appsDirectory, appEntry.name, ".next");
		for (const nextEntry of await readDirectorySafely(nextDirectory)) {
			if (
				nextEntry.isDirectory() &&
				nextEntry.name.startsWith("dev-stale")
			) {
				targets.push(path.join(nextDirectory, nextEntry.name));
			}
		}
	}

	return targets;
}

async function pruneTurboCache() {
	const cacheDirectory = resolveFromRoot(".turbo", "cache");
	assertSafeTarget(cacheDirectory);
	const maximumGiB = Number(process.env.SHAHRAH_TURBO_CACHE_MAX_GIB ?? "5");
	if (!Number.isFinite(maximumGiB) || maximumGiB <= 0) {
		fail("SHAHRAH_TURBO_CACHE_MAX_GIB must be a positive number.");
	}

	const maximumBytes = maximumGiB * 1024 ** 3;
	const groups = new Map();

	for (const entry of await readDirectorySafely(cacheDirectory)) {
		if (!entry.isFile()) {
			continue;
		}

		const match = entry.name.match(
			/^([a-f0-9]+)(?:-manifest\.json|\.tar\.zst)$/,
		);
		if (!match) {
			continue;
		}

		const filePath = path.join(cacheDirectory, entry.name);
		const info = await lstat(filePath);
		const group = groups.get(match[1]) ?? {
			files: [],
			size: 0,
			modifiedAt: 0,
		};
		group.files.push(filePath);
		group.size += info.size;
		group.modifiedAt = Math.max(group.modifiedAt, info.mtimeMs);
		groups.set(match[1], group);
	}

	const orderedGroups = [...groups.values()].sort(
		(left, right) => left.modifiedAt - right.modifiedAt,
	);
	let totalBytes = orderedGroups.reduce((sum, group) => sum + group.size, 0);
	let reclaimedBytes = 0;

	for (const group of orderedGroups) {
		if (totalBytes <= maximumBytes) {
			break;
		}

		for (const filePath of group.files) {
			assertSafeTarget(filePath);
			console.log(
				`${dryRun ? "Would remove" : "Removing"} ${relativePath(filePath)}`,
			);
			if (!dryRun) {
				await rm(filePath, { force: true });
			}
		}
		totalBytes -= group.size;
		reclaimedBytes += group.size;
	}

	console.log(`Turbo cache limit: ${maximumGiB} GiB`);
	console.log(
		`${dryRun ? "Potential reclaim" : "Reclaimed"}: ${formatBytes(reclaimedBytes)}`,
	);
	console.log(
		`${dryRun ? "Projected cache size" : "Cache size"}: ${formatBytes(totalBytes)}`,
	);
}

async function verifyRepositoryRoot() {
	const packageJson = JSON.parse(
		await readFile(resolveFromRoot("package.json"), "utf8"),
	);
	if (packageJson.name !== "shahrah") {
		fail(`Refusing to clean unexpected repository: ${repositoryRoot}`);
	}
}

function assertSafeTarget(target) {
	const resolvedTarget = path.resolve(target);
	const relativeTarget = path.relative(repositoryRoot, resolvedTarget);
	if (
		!relativeTarget ||
		relativeTarget === "." ||
		relativeTarget.startsWith(`..${path.sep}`) ||
		path.isAbsolute(relativeTarget)
	) {
		fail(`Refusing unsafe target: ${resolvedTarget}`);
	}
}

async function getPathSize(target) {
	let info;
	try {
		info = await lstat(target);
	} catch (error) {
		if (error.code === "ENOENT") {
			return null;
		}
		throw error;
	}

	if (!info.isDirectory() || info.isSymbolicLink()) {
		return info.size;
	}

	let total = 0;
	for (const entry of await readdir(target, { withFileTypes: true })) {
		total += (await getPathSize(path.join(target, entry.name))) ?? 0;
	}
	return total;
}

async function getPathStats(target) {
	let info;
	try {
		info = await lstat(target);
	} catch (error) {
		if (error.code === "ENOENT") {
			return null;
		}
		throw error;
	}

	if (!info.isDirectory() || info.isSymbolicLink()) {
		return { size: info.size, latestModifiedAt: info.mtimeMs };
	}

	let size = 0;
	let latestModifiedAt = info.mtimeMs;
	for (const entry of await readdir(target, { withFileTypes: true })) {
		const childStats = await getPathStats(path.join(target, entry.name));
		if (childStats === null) {
			continue;
		}
		size += childStats.size;
		latestModifiedAt = Math.max(
			latestModifiedAt,
			childStats.latestModifiedAt,
		);
	}
	return { size, latestModifiedAt };
}

function createSelection(targets, managedTargets, reasons) {
	return {
		targets: uniquePaths(targets),
		managedTargets: uniquePaths(managedTargets),
		reasons,
	};
}

function readPositiveNumber(name, fallback) {
	const value = Number(process.env[name] ?? fallback);
	if (!Number.isFinite(value) || value <= 0) {
		fail(`${name} must be a positive number.`);
	}
	return value;
}

async function readDirectorySafely(directory) {
	try {
		return await readdir(directory, { withFileTypes: true });
	} catch (error) {
		if (error.code === "ENOENT") {
			return [];
		}
		throw error;
	}
}

function uniquePaths(paths) {
	return [...new Set(paths.map((target) => path.resolve(target)))];
}

function resolveFromRoot(...segments) {
	return path.resolve(repositoryRoot, ...segments);
}

function relativePath(target) {
	return path.relative(repositoryRoot, target) || ".";
}

function formatBytes(bytes) {
	if (bytes < 1024 ** 2) {
		return `${(bytes / 1024).toFixed(1)} KiB`;
	}
	if (bytes < 1024 ** 3) {
		return `${(bytes / 1024 ** 2).toFixed(1)} MiB`;
	}
	return `${(bytes / 1024 ** 3).toFixed(2)} GiB`;
}

function fail(message) {
	console.error(message);
	process.exit(1);
}
