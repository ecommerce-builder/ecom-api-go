CREATE VIEW category_leaf AS
SELECT *
FROM category
WHERE lft = rgt - 1 ORDER BY lft ASC;
