-- CreateEnum
CREATE TYPE "ModelMode" AS ENUM (
  'completion',
  'chat',
  'embedding',
  'audio_transcription',
  'image',
  'video',
  'rerank',
  'text_to_speech',
  'realtime',
  'responses',
  'moderation',
  'unknown'
);

-- AlterTable
ALTER TABLE "Model"
ADD COLUMN "mode" "ModelMode" NOT NULL DEFAULT 'unknown';
