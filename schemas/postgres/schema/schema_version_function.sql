create function schema_version()
  returns text
as
$$
  select 'v0.58.1' || '';
$$
language sql;
