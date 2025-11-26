-- Save or update a value in plugin storage
-- Parameters: plugin_name, storage_key, storage_value, updated_at
-- Uses MERGE statement for MSSQL upsert semantics

MERGE INTO plugin_storage AS target
USING (SELECT @p1 AS plugin_name, @p2 AS storage_key, @p3 AS storage_value, @p4 AS updated_at) AS source
ON target.plugin_name = source.plugin_name AND target.storage_key = source.storage_key
WHEN MATCHED THEN
    UPDATE SET
        storage_value = source.storage_value,
        updated_at = source.updated_at
WHEN NOT MATCHED THEN
    INSERT (plugin_name, storage_key, storage_value, updated_at)
    VALUES (source.plugin_name, source.storage_key, source.storage_value, source.updated_at);
