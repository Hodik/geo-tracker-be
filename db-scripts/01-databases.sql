SELECT 'CREATE DATABASE geo_tracker'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'geo_tracker')\gexec