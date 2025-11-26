-- Create plugin_events table for MSSQL (SQL Server)
-- Stores plugin event subscriptions and publications
-- Tracks which plugins publish and subscribe to events
-- Foreign keys reference plugins table

IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[plugin_events]') AND type in (N'U'))
BEGIN
    CREATE TABLE plugin_events (
        id BIGINT IDENTITY(1,1) PRIMARY KEY,
        event_name NVARCHAR(255) NOT NULL,
        publisher_plugin NVARCHAR(255) NOT NULL,
        subscriber_plugin NVARCHAR(255) NOT NULL,
        created_at DATETIME2 DEFAULT GETDATE(),
        FOREIGN KEY (publisher_plugin) REFERENCES plugins(name) ON DELETE CASCADE,
        FOREIGN KEY (subscriber_plugin) REFERENCES plugins(name) ON DELETE CASCADE
    );
END;
