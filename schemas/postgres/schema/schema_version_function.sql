create function schema_version()
  returns text
as
$$
  select 'v0.63.0' || '';
$$
language sql;
