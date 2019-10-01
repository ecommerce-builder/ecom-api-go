create function schema_version()
  returns text
as
$$
  select 'v0.60.0' || '';
$$
language sql;
