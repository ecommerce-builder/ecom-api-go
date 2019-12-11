create function schema_version()
  returns text
as
$$
  select 'v0.64.0' || '';
$$
language sql;
