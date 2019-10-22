create function schema_version()
  returns text
as
$$
  select 'v0.61.3' || '';
$$
language sql;
