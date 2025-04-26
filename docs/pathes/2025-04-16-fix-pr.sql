SELECT column_name
FROM information_schema.key_column_usage
WHERE table_name = 'ludomans' AND constraint_name = 'PRIMARY';
ALTER TABLE "ludomans" ADD PRIMARY KEY ("ludomanId");