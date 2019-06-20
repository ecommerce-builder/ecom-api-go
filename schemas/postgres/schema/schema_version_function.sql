create function schema_version()
  returns text
as
$$
  select 'v0.47.1' || '';
$$
language sql;
