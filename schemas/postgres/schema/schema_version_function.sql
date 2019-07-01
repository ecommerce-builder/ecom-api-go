create function schema_version()
  returns text
as
$$
  select 'v0.50.0' || '';
$$
language sql;
