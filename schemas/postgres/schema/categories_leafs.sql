CREATE VIEW categories_leafs AS
SELECT *
FROM categories
WHERE lft = rgt - 1 ORDER BY lft ASC;
