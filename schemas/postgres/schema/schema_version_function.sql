create function schema_version()
  returns text
as
$$
  select 'v0.49.1' || '';
$$
language sql;
