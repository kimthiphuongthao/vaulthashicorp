-- Create legacy schema/table used by the SQL updater demo

IF DB_ID('legacy') IS NULL
BEGIN
    CREATE DATABASE legacy;
END
GO

USE legacy;
GO

IF OBJECT_ID('dbo.users', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.users (
        username NVARCHAR(255) NOT NULL PRIMARY KEY,
        password_hash NVARCHAR(MAX) NOT NULL DEFAULT('')
    );
END
GO

IF NOT EXISTS (SELECT 1 FROM dbo.users WHERE username = 'alice')
    INSERT INTO dbo.users (username, password_hash) VALUES ('alice', '');
IF NOT EXISTS (SELECT 1 FROM dbo.users WHERE username = 'bob')
    INSERT INTO dbo.users (username, password_hash) VALUES ('bob', '');
IF NOT EXISTS (SELECT 1 FROM dbo.users WHERE username = 'charlie')
    INSERT INTO dbo.users (username, password_hash) VALUES ('charlie', '');
GO
