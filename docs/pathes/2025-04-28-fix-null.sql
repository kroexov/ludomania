UPDATE ludomans
SET "totalLost" = 0
WHERE "totalLost" IS NULL;

UPDATE ludomans
SET "totalWon" = 0
WHERE "totalWon" IS NULL;

ALTER TABLE ludomans
ALTER COLUMN "totalLost" SET DEFAULT 0,
ALTER COLUMN "totalWon"  SET DEFAULT 0;

ALTER TABLE ludomans
ALTER COLUMN "totalLost" SET NOT NULL,
ALTER COLUMN "totalWon"  SET NOT NULL;

