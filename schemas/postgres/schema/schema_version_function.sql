create function schema_version()
  returns text
as
$$
  select 'v0.58.0' || '';
$$
language sql;
