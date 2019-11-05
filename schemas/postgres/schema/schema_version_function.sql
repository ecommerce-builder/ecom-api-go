create function schema_version()
  returns text
as
$$
  select 'v0.62.2' || '';
$$
language sql;
