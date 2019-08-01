create function schema_version()
  returns text
as
$$
  select 'v0.56.0' || '';
$$
language sql;
