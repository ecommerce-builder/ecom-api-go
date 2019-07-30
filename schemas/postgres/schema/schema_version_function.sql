create function schema_version()
  returns text
as
$$
  select 'v0.54.0' || '';
$$
language sql;
