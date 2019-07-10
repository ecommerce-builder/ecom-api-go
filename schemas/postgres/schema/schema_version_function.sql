create function schema_version()
  returns text
as
$$
  select 'v0.52.0' || '';
$$
language sql;
