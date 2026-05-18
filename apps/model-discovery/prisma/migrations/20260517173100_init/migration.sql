-- CreateSchema
CREATE SCHEMA IF NOT EXISTS "public";

-- CreateEnum
CREATE TYPE "ProviderStatus" AS ENUM ('active', 'preview', 'deprecated', 'unavailable');

-- CreateEnum
CREATE TYPE "ModelStatus" AS ENUM ('active', 'preview', 'deprecated', 'unavailable');

-- CreateEnum
CREATE TYPE "LifecycleStatus" AS ENUM ('draft', 'review', 'published', 'deprecated', 'archived');

-- CreateEnum
CREATE TYPE "ModalityDirectionType" AS ENUM ('input', 'output', 'both');

-- CreateTable
CREATE TABLE "Provider" (
    "id" TEXT NOT NULL,
    "slug" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "websiteUrl" TEXT,
    "docsUrl" TEXT,
    "logoUrl" TEXT,
    "status" "ProviderStatus" NOT NULL DEFAULT 'active',
    "privacyPolicyUrl" TEXT,
    "loggingPolicySummary" TEXT,
    "supportsByok" BOOLEAN NOT NULL DEFAULT false,
    "supportsRegionalRouting" BOOLEAN NOT NULL DEFAULT false,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "Provider_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Model" (
    "id" TEXT NOT NULL,
    "providerId" TEXT NOT NULL,
    "canonicalId" TEXT NOT NULL,
    "slug" TEXT NOT NULL,
    "displayName" TEXT NOT NULL,
    "shortDescription" TEXT,
    "longDescription" TEXT,
    "status" "ModelStatus" NOT NULL DEFAULT 'active',
    "lifecycleStatus" "LifecycleStatus" NOT NULL DEFAULT 'draft',
    "categoryId" TEXT,
    "contextWindow" INTEGER,
    "maxOutputTokens" INTEGER,
    "knowledgeCutoff" TIMESTAMP(3),
    "releaseDate" TIMESTAMP(3),
    "deprecatedAt" TIMESTAMP(3),
    "openSource" BOOLEAN NOT NULL DEFAULT false,
    "free" BOOLEAN NOT NULL DEFAULT false,
    "tokenizer" TEXT,
    "license" TEXT,
    "huggingFaceUrl" TEXT,
    "modelCardUrl" TEXT,
    "officialDocsUrl" TEXT,
    "popularityScore" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "Model_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ModelVersion" (
    "id" TEXT NOT NULL,
    "modelId" TEXT NOT NULL,
    "version" TEXT NOT NULL,
    "releaseDate" TIMESTAMP(3),
    "notes" TEXT,
    "status" "ModelStatus" NOT NULL DEFAULT 'active',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "ModelVersion_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ModelPricing" (
    "id" TEXT NOT NULL,
    "modelId" TEXT NOT NULL,
    "regionId" TEXT,
    "currency" TEXT NOT NULL DEFAULT 'USD',
    "inputPricePerMillionTokens" DOUBLE PRECISION,
    "outputPricePerMillionTokens" DOUBLE PRECISION,
    "cachedInputPricePerMillionTokens" DOUBLE PRECISION,
    "imageInputPrice" DOUBLE PRECISION,
    "imageOutputPrice" DOUBLE PRECISION,
    "audioInputPrice" DOUBLE PRECISION,
    "audioOutputPrice" DOUBLE PRECISION,
    "requestPrice" DOUBLE PRECISION,
    "effectiveFrom" TIMESTAMP(3),
    "effectiveTo" TIMESTAMP(3),
    "sourceUrl" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "ModelPricing_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Capability" (
    "id" TEXT NOT NULL,
    "slug" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,

    CONSTRAINT "Capability_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ModelCapability" (
    "modelId" TEXT NOT NULL,
    "capabilityId" TEXT NOT NULL,

    CONSTRAINT "ModelCapability_pkey" PRIMARY KEY ("modelId","capabilityId")
);

-- CreateTable
CREATE TABLE "Modality" (
    "id" TEXT NOT NULL,
    "slug" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "directionType" "ModalityDirectionType" NOT NULL,

    CONSTRAINT "Modality_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ModelInputModality" (
    "modelId" TEXT NOT NULL,
    "modalityId" TEXT NOT NULL,

    CONSTRAINT "ModelInputModality_pkey" PRIMARY KEY ("modelId","modalityId")
);

-- CreateTable
CREATE TABLE "ModelOutputModality" (
    "modelId" TEXT NOT NULL,
    "modalityId" TEXT NOT NULL,

    CONSTRAINT "ModelOutputModality_pkey" PRIMARY KEY ("modelId","modalityId")
);

-- CreateTable
CREATE TABLE "SupportedParameter" (
    "id" TEXT NOT NULL,
    "slug" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,

    CONSTRAINT "SupportedParameter_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ModelSupportedParameter" (
    "modelId" TEXT NOT NULL,
    "supportedParameterId" TEXT NOT NULL,

    CONSTRAINT "ModelSupportedParameter_pkey" PRIMARY KEY ("modelId","supportedParameterId")
);

-- CreateTable
CREATE TABLE "Category" (
    "id" TEXT NOT NULL,
    "slug" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,

    CONSTRAINT "Category_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Region" (
    "id" TEXT NOT NULL,
    "slug" TEXT NOT NULL,
    "name" TEXT NOT NULL,

    CONSTRAINT "Region_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ModelRegion" (
    "modelId" TEXT NOT NULL,
    "regionId" TEXT NOT NULL,

    CONSTRAINT "ModelRegion_pkey" PRIMARY KEY ("modelId","regionId")
);

-- CreateTable
CREATE TABLE "Benchmark" (
    "id" TEXT NOT NULL,
    "slug" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "maxScore" DOUBLE PRECISION,
    "scoreType" TEXT,

    CONSTRAINT "Benchmark_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ModelBenchmark" (
    "id" TEXT NOT NULL,
    "modelId" TEXT NOT NULL,
    "benchmarkId" TEXT NOT NULL,
    "score" DOUBLE PRECISION NOT NULL,
    "percentile" DOUBLE PRECISION,
    "sourceUrl" TEXT,
    "evaluatedAt" TIMESTAMP(3),
    "notes" TEXT,

    CONSTRAINT "ModelBenchmark_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "SourceReference" (
    "id" TEXT NOT NULL,
    "modelId" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "url" TEXT NOT NULL,
    "sourceType" TEXT,
    "fetchedAt" TIMESTAMP(3),
    "notes" TEXT,

    CONSTRAINT "SourceReference_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ModelChangelog" (
    "id" TEXT NOT NULL,
    "modelId" TEXT NOT NULL,
    "changeType" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "description" TEXT,
    "changedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "createdBy" TEXT,

    CONSTRAINT "ModelChangelog_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Tag" (
    "id" TEXT NOT NULL,
    "slug" TEXT NOT NULL,
    "name" TEXT NOT NULL,

    CONSTRAINT "Tag_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ModelTag" (
    "modelId" TEXT NOT NULL,
    "tagId" TEXT NOT NULL,

    CONSTRAINT "ModelTag_pkey" PRIMARY KEY ("modelId","tagId")
);

-- CreateTable
CREATE TABLE "AuditLog" (
    "id" TEXT NOT NULL,
    "entityType" TEXT NOT NULL,
    "entityId" TEXT NOT NULL,
    "action" TEXT NOT NULL,
    "beforeJson" JSONB,
    "afterJson" JSONB,
    "actorId" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "AuditLog_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "Provider_slug_key" ON "Provider"("slug");

-- CreateIndex
CREATE INDEX "Provider_slug_idx" ON "Provider"("slug");

-- CreateIndex
CREATE UNIQUE INDEX "Model_canonicalId_key" ON "Model"("canonicalId");

-- CreateIndex
CREATE INDEX "Model_providerId_idx" ON "Model"("providerId");

-- CreateIndex
CREATE INDEX "Model_status_idx" ON "Model"("status");

-- CreateIndex
CREATE INDEX "Model_contextWindow_idx" ON "Model"("contextWindow");

-- CreateIndex
CREATE INDEX "Model_maxOutputTokens_idx" ON "Model"("maxOutputTokens");

-- CreateIndex
CREATE INDEX "Model_slug_idx" ON "Model"("slug");

-- CreateIndex
CREATE UNIQUE INDEX "Model_providerId_slug_key" ON "Model"("providerId", "slug");

-- CreateIndex
CREATE INDEX "ModelPricing_modelId_idx" ON "ModelPricing"("modelId");

-- CreateIndex
CREATE INDEX "ModelPricing_effectiveFrom_effectiveTo_idx" ON "ModelPricing"("effectiveFrom", "effectiveTo");

-- CreateIndex
CREATE UNIQUE INDEX "Capability_slug_key" ON "Capability"("slug");

-- CreateIndex
CREATE UNIQUE INDEX "Modality_slug_key" ON "Modality"("slug");

-- CreateIndex
CREATE UNIQUE INDEX "SupportedParameter_slug_key" ON "SupportedParameter"("slug");

-- CreateIndex
CREATE UNIQUE INDEX "Category_slug_key" ON "Category"("slug");

-- CreateIndex
CREATE UNIQUE INDEX "Region_slug_key" ON "Region"("slug");

-- CreateIndex
CREATE UNIQUE INDEX "Benchmark_slug_key" ON "Benchmark"("slug");

-- CreateIndex
CREATE INDEX "ModelBenchmark_modelId_benchmarkId_idx" ON "ModelBenchmark"("modelId", "benchmarkId");

-- CreateIndex
CREATE UNIQUE INDEX "Tag_slug_key" ON "Tag"("slug");

-- CreateIndex
CREATE INDEX "AuditLog_entityType_entityId_idx" ON "AuditLog"("entityType", "entityId");

-- AddForeignKey
ALTER TABLE "Model" ADD CONSTRAINT "Model_providerId_fkey" FOREIGN KEY ("providerId") REFERENCES "Provider"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "Model" ADD CONSTRAINT "Model_categoryId_fkey" FOREIGN KEY ("categoryId") REFERENCES "Category"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelVersion" ADD CONSTRAINT "ModelVersion_modelId_fkey" FOREIGN KEY ("modelId") REFERENCES "Model"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelPricing" ADD CONSTRAINT "ModelPricing_modelId_fkey" FOREIGN KEY ("modelId") REFERENCES "Model"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelPricing" ADD CONSTRAINT "ModelPricing_regionId_fkey" FOREIGN KEY ("regionId") REFERENCES "Region"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelCapability" ADD CONSTRAINT "ModelCapability_modelId_fkey" FOREIGN KEY ("modelId") REFERENCES "Model"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelCapability" ADD CONSTRAINT "ModelCapability_capabilityId_fkey" FOREIGN KEY ("capabilityId") REFERENCES "Capability"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelInputModality" ADD CONSTRAINT "ModelInputModality_modelId_fkey" FOREIGN KEY ("modelId") REFERENCES "Model"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelInputModality" ADD CONSTRAINT "ModelInputModality_modalityId_fkey" FOREIGN KEY ("modalityId") REFERENCES "Modality"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelOutputModality" ADD CONSTRAINT "ModelOutputModality_modelId_fkey" FOREIGN KEY ("modelId") REFERENCES "Model"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelOutputModality" ADD CONSTRAINT "ModelOutputModality_modalityId_fkey" FOREIGN KEY ("modalityId") REFERENCES "Modality"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelSupportedParameter" ADD CONSTRAINT "ModelSupportedParameter_modelId_fkey" FOREIGN KEY ("modelId") REFERENCES "Model"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelSupportedParameter" ADD CONSTRAINT "ModelSupportedParameter_supportedParameterId_fkey" FOREIGN KEY ("supportedParameterId") REFERENCES "SupportedParameter"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelRegion" ADD CONSTRAINT "ModelRegion_modelId_fkey" FOREIGN KEY ("modelId") REFERENCES "Model"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelRegion" ADD CONSTRAINT "ModelRegion_regionId_fkey" FOREIGN KEY ("regionId") REFERENCES "Region"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelBenchmark" ADD CONSTRAINT "ModelBenchmark_modelId_fkey" FOREIGN KEY ("modelId") REFERENCES "Model"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelBenchmark" ADD CONSTRAINT "ModelBenchmark_benchmarkId_fkey" FOREIGN KEY ("benchmarkId") REFERENCES "Benchmark"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "SourceReference" ADD CONSTRAINT "SourceReference_modelId_fkey" FOREIGN KEY ("modelId") REFERENCES "Model"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelChangelog" ADD CONSTRAINT "ModelChangelog_modelId_fkey" FOREIGN KEY ("modelId") REFERENCES "Model"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelTag" ADD CONSTRAINT "ModelTag_modelId_fkey" FOREIGN KEY ("modelId") REFERENCES "Model"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "ModelTag" ADD CONSTRAINT "ModelTag_tagId_fkey" FOREIGN KEY ("tagId") REFERENCES "Tag"("id") ON DELETE CASCADE ON UPDATE CASCADE;

