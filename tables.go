package main

// TABLES_SQL defines the main database tables
// and trigger functions.
var TABLES_SQL = `

    CREATE TABLE IF NOT EXISTS assets (
        path VARCHAR NOT NULL PRIMARY KEY,
        content VARCHAR,
        type VARCHAR,
        create_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        update_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        UNIQUE(path)
    );

    CREATE TRIGGER IF NOT EXISTS assets__update
        AFTER
        UPDATE
        ON assets
        FOR EACH ROW
    BEGIN
        UPDATE assets SET update_at=CURRENT_TIMESTAMP WHERE path=OLD.path;
    END;



    CREATE TABLE IF NOT EXISTS users (
        username TEXT,
        password TEXT,
        create_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        update_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

`
