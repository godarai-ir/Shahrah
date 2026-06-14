-- Convert model invocation mode from one enum value to multiple string values.
ALTER TABLE "Model"
ALTER COLUMN "mode" DROP DEFAULT;

ALTER TABLE "Model"
ALTER COLUMN "mode" TYPE TEXT[]
USING ARRAY[("mode")::TEXT];

ALTER TABLE "Model"
ALTER COLUMN "mode" SET DEFAULT ARRAY['unknown']::TEXT[];

-- Direction is represented by ModelInputModality and ModelOutputModality joins.
ALTER TABLE "Modality"
DROP COLUMN "directionType";

DROP TYPE IF EXISTS "ModalityDirectionType";
DROP TYPE IF EXISTS "ModelMode";
