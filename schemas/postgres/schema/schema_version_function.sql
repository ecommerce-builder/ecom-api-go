create function schema_version()
  returns text
as
$$
  select 'v0.55.0' || '';
$$
language sql;
