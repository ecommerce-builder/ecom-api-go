create function schema_version()
  returns text
as
$$
  select 'v0.60.1' || '';
$$
language sql;
