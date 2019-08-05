create function schema_version()
  returns text
as
$$
  select 'v0.57.0' || '';
$$
language sql;
