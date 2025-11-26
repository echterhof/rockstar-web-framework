-- Save or update a tenant in the database
-- Parameters: id, name, hosts (JSON), config (JSON), is_active, created_at, updated_at, max_users, max_storage, max_requests
-- Uses MERGE statement for MSSQL upsert semantics

MERGE INTO tenants AS target
USING (SELECT ? AS id, ? AS name, ? AS hosts, ? AS config, ? AS is_active, 
              ? AS created_at, ? AS updated_at, ? AS max_users, ? AS max_storage, ? AS max_requests) AS source
ON target.id = source.id
WHEN MATCHED THEN
    UPDATE SET
        name = source.name,
        hosts = source.hosts,
        config = source.config,
        is_active = source.is_active,
        updated_at = source.updated_at,
        max_users = source.max_users,
        max_storage = source.max_storage,
        max_requests = source.max_requests
WHEN NOT MATCHED THEN
    INSERT (id, name, hosts, config, is_active, created_at, updated_at, max_users, max_storage, max_requests)
    VALUES (source.id, source.name, source.hosts, source.config, source.is_active, 
            source.created_at, source.updated_at, source.max_users, source.max_storage, source.max_requests);
