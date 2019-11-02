create function schema_version()
  returns text
as
$$
  select 'v0.62.0' || '';
$$
language sql;
