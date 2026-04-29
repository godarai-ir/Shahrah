import fs from "node:fs";
import path from "node:path";
import { DatabaseSync } from "node:sqlite";

export type SignupSubmission = {
  mobile: string;
  teamSize: string;
  usageType: string;
  activityField: string;
  features: string[];
};

type SignupRow = {
  id: number;
  mobile: string;
  team_size: string;
  usage_type: string;
  activity_field: string;
  features_json: string;
  created_at: string;
};

const dataDir = path.join(process.cwd(), "data");
const dbPath = path.join(dataDir, "signup.sqlite");

let db: DatabaseSync | null = null;

function getDb() {
  if (db) return db;

  if (!fs.existsSync(dataDir)) {
    fs.mkdirSync(dataDir, { recursive: true });
  }

  db = new DatabaseSync(dbPath);
  db.exec(`
    CREATE TABLE IF NOT EXISTS signup_submissions (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      mobile TEXT NOT NULL,
      team_size TEXT NOT NULL,
      usage_type TEXT NOT NULL,
      activity_field TEXT NOT NULL,
      features_json TEXT NOT NULL,
      created_at TEXT NOT NULL DEFAULT (datetime('now'))
    );
  `);

  return db;
}

export async function insertSignupSubmission(input: SignupSubmission) {
  const database = getDb();
  const stmt = database.prepare(`
    INSERT INTO signup_submissions (
      mobile, team_size, usage_type, activity_field, features_json
    ) VALUES (?, ?, ?, ?, ?);
  `);

  stmt.run(
    input.mobile,
    input.teamSize,
    input.usageType,
    input.activityField,
    JSON.stringify(input.features ?? []),
  );

  const idRow = database
    .prepare("SELECT last_insert_rowid() AS id;")
    .get() as { id: number };
  return Number(idRow.id);
}

export async function listSignupSubmissions(): Promise<Array<
  SignupRow & { features: string[] }
>> {
  const database = getDb();
  const rows = database.prepare(`
    SELECT id, mobile, team_size, usage_type, activity_field, features_json, created_at
    FROM signup_submissions
    ORDER BY id DESC;
  `).all() as SignupRow[];

  return rows.map((row) => ({
    ...row,
    features: safeParseFeatures(row.features_json),
  }));
}

function safeParseFeatures(raw: string): string[] {
  try {
    const parsed = JSON.parse(raw);
    if (Array.isArray(parsed)) {
      return parsed.map((item) => String(item));
    }
    return [];
  } catch {
    return [];
  }
}

