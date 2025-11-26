-- Create plugin_events table for MySQL
-- Stores plugin event subscriptions and publications
-- Tracks which plugins publish and subscribe to events
-- Foreign keys reference plugins table

CREATE TABLE IF NOT EXISTS plugin_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_name VARCHAR(255) NOT NULL,
    publisher_plugin VARCHAR(255) NOT NULL,
    subscriber_plugin VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (publisher_plugin) REFERENCES plugins(name) ON DELETE CASCADE,
    FOREIGN KEY (subscriber_plugin) REFERENCES plugins(name) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
