create function schema_version()
  returns text
as
$$
  select 'v0.62.1' || '';
$$
language sql;
