create function schema_version()
  returns text
as
$$
  select 'v0.61.0' || '';
$$
language sql;
