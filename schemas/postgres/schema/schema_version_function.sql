create function schema_version()
  returns text
as
$$
  select 'v0.49.4' || '';
$$
language sql;
