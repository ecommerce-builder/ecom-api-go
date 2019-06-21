create function schema_version()
  returns text
as
$$
  select 'v0.48.0' || '';
$$
language sql;
