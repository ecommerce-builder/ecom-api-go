create function schema_version()
  returns text
as
$$
  select 'v0.47.0' || '';
$$
language sql;
