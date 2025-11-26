-- Save or update an access token in the database (MSSQL)
-- Parameters: @p1=token, @p2=user_id, @p3=tenant_id, @p4=scopes (JSON), @p5=expires_at, @p6=created_at
-- Uses MERGE statement for MSSQL upsert semantics
-- MERGE is the standard way to perform upsert operations in SQL Server

MERGE INTO access_tokens AS target
USING (SELECT ? AS token, ? AS user_id, ? AS tenant_id, ? AS scopes, ? AS expires_at, ? AS created_at) AS source
ON target.token = source.token
WHEN MATCHED THEN
    UPDATE SET
        user_id = source.user_id,
        tenant_id = source.tenant_id,
        scopes = source.scopes,
        expires_at = source.expires_at
WHEN NOT MATCHED THEN
    INSERT (token, user_id, tenant_id, scopes, expires_at, created_at)
    VALUES (source.token, source.user_id, source.tenant_id, source.scopes, source.expires_at, source.created_at);
