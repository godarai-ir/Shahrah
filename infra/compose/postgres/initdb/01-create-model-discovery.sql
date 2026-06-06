SELECT 'CREATE DATABASE model_discovery'
WHERE NOT EXISTS (
  SELECT FROM pg_database WHERE datname = 'model_discovery'
)\gexec
