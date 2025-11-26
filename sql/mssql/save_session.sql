-- Save or update a session in the database (MSSQL)
-- Parameters: @p1=id, @p2=user_id, @p3=tenant_id, @p4=data (JSON), @p5=expires_at, 
--            @p6=created_at, @p7=updated_at, @p8=ip_address, @p9=user_agent
-- Uses MERGE statement for MSSQL upsert semantics

MERGE INTO sessions AS target
USING (SELECT ? AS id, ? AS user_id, ? AS tenant_id, ? AS data, ? AS expires_at, 
              ? AS created_at, ? AS updated_at, ? AS ip_address, ? AS user_agent) AS source
ON (target.id = source.id)
WHEN MATCHED THEN
    UPDATE SET 
        user_id = source.user_id,
        tenant_id = source.tenant_id,
        data = source.data,
        expires_at = source.expires_at,
        updated_at = source.updated_at,
        ip_address = source.ip_address,
        user_agent = source.user_agent
WHEN NOT MATCHED THEN
    INSERT (id, user_id, tenant_id, data, expires_at, created_at, updated_at, ip_address, user_agent)
    VALUES (source.id, source.user_id, source.tenant_id, source.data, source.expires_at, 
            source.created_at, source.updated_at, source.ip_address, source.user_agent);
