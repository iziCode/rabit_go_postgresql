
-- DROP TABLE IF EXISTS "queue";
-- DROP SEQUENCE IF EXISTS queue_id_seq;
CREATE SEQUENCE IF NOT EXISTS queue_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1;

CREATE TABLE IF NOT EXISTS "public"."queue" (
    "id" integer DEFAULT nextval('queue_id_seq') NOT NULL,
    "msg" text NOT NULL
) WITH (oids = false);

INSERT INTO "queue" ("id", "msg") VALUES
(1,	'Test data');

