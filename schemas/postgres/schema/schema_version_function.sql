create function schema_version()
  returns text
as
$$
  select 'v0.59.0' || '';
$$
language sql;
