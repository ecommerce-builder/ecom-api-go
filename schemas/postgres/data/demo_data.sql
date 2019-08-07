-- customers and addresses
-- 1
INSERT INTO customers (uid, role, email, firstname, lastname)
  VALUES ('uid1', 'customer', 'joe@example.com', 'Joe', 'Blogs');

-- address: id=1
INSERT INTO addresses (typ, customer_id, contact_name, addr1, city, postcode, country)
  VALUES('billing', 1, 'Joe Blogs', '4524 Mulberry Avenue', 'LittleRock', '72209', 'US');


-- 2
INSERT INTO customers (uid, role, email, firstname, lastname)
  VALUES ('uid2', 'customer', 'sammy@example.com', 'Sammy', 'Peterson');

-- address: id=2
INSERT INTO addresses (typ, customer_id, contact_name, addr1, city, postcode, country)
  VALUES('billing', 2, 'Sammy Peterson', '138 Ermin Street', 'Wrentham', 'NR34 9TT', 'UK');


-- 3
INSERT INTO customers (uid, role, email, firstname, lastname)
  VALUES ('uid3', 'customer', 'faith@example.com', 'Faith', 'Bowman');

-- address: id=3
INSERT INTO addresses (typ, customer_id, contact_name, addr1, city, postcode, country)
  VALUES('shipping', 3, 'Faith Bowman', '18 Pier Road', 'Statham', 'WA13 3DW', 'UK');

-- address: id=4
INSERT INTO addresses (typ, customer_id, contact_name, addr1, city, postcode, country)
  VALUES('shipping', 3, 'Faith Bowman', '115 Spilman Street', 'Gossops Green', 'RH11 9SP', 'UK');

-- address: id=5
INSERT INTO addresses (typ, customer_id, contact_name, addr1, city, postcode, country)
  VALUES('shipping', 3, 'Faith Bowman', '43 Shannon Way', 'Chipping Campden', 'GL55 9XZ', 'UK');

-- address: id=6
INSERT INTO addresses (typ, customer_id, contact_name, addr1, city, postcode, country)
  VALUES('billing', 3, 'Faith Bowman', '38 Walden Road', 'Greenburn', 'DD5 8AU','UK');

-- address: id=7
INSERT INTO addresses (typ, customer_id, contact_name, addr1, city, postcode, country)
  VALUES('billing', 3, 'Faith Bowman', '99  Wrexham Rd', 'Faceby', 'TS9 4QL', 'UK');


-- 4
INSERT INTO customers (uid, role, email, firstname, lastname)
  VALUES ('uid4', 'customer', 'clifton@example.com', 'Clifton', 'Delgado');

-- address: id=8
INSERT INTO addresses (typ, customer_id, contact_name, addr1, city, postcode, country)
  VALUES('billing', 4, 'Clifton Delgado', '131 Caxton Place', 'Byfield', 'NN11 7FN', 'UK');


-- 5
INSERT INTO customers (uid, role, email, firstname, lastname)
  VALUES ('uid5', 'customer', 'bernadette@example.com', 'Bernadette', 'Graham');

-- address: id=9
INSERT INTO addresses (typ, customer_id, contact_name, addr1, city, postcode, country)
  VALUES('shipping',5, 'Bernadette Graham', '89 Cubbine Road', 'Southburracoppin', '6421', 'AU');

-- address: id=10
INSERT INTO addresses (typ, customer_id, contact_name, addr1, city, postcode, country)
  VALUES('billing',5, 'Bernadette Graham', '38 Porana Place', 'Woolgorong', '6630', 'AU');


-- categories
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('a', 'a', 'ประเภท A', 1, 28, 0, now(), now());
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('b', 'a/b', 'ประเภท B', 2, 5, 1, now(), now());
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('e', 'a/b/e', 'ประเภท E', 3, 4, 2, now(), now());
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('c', 'a/c', 'ประเภท C', 6, 19, 1, now(), now());
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('f', 'a/c/f', 'ประเภท F', 7, 16, 2, now(), now());
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('i', 'a/c/f/i', 'ประเภท I', 8, 9, 3, now(), now());
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('j', 'a/c/f/j', 'ประเภท J', 10, 15, 3, now(), now());
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('m', 'a/c/f/j/m', 'ประเภท M', 11, 12, 4, now(), now());
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('n', 'a/c/f/j/n', 'ประเภท N', 13, 14, 4, now(), now());
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('g', 'a/c/g', 'ประเภท G', 17, 18, 2, now(), now());
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('d', 'a/d', 'ประเภท D', 20, 27, 1, now(), now());
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('h', 'a/d/h', 'ประเภท H', 21, 26, 2, now(), now());
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('k', 'a/d/h/k', 'ประเภท K', 22, 23, 3, now(), now());
INSERT INTO categories (segment, path, name, lft, rgt, depth, created, modified)
  VALUES ('l', 'a/d/h/l', 'ประเภท L', 24, 25, 3, now(), now());

-- orders, order_items and payments
-- 100001
INSERT INTO orders (customer_id, ship_tb, billing, shipping, total_ex_vat, vat_total, total_inc_vat)
  VALUES(1, true, '{"address_id": 1, "contact_name": "Joe Blogs", "addr1": "4524 Mulberry Avenue",
            "city": "LittleRock", "postcode": "72209", "country": "US"}', NULL, 9999, 2000, 11999);
INSERT INTO order_items (order_id, sku, name, qty, unit_price, currency, discount, tax_code, vat)
  VALUES(100001, 'FRIDGE', 'Luxuary Fridge', 1, 9554, 'GBP', null, 'T20', 1911);
INSERT INTO order_items (order_id, sku, name, qty, unit_price, currency, discount, tax_code, vat)
  VALUES(100001, 'WATER-SKU', 'Water Bottle', 1, 445, 'GBP', null, 'T20', 89);

-- 100002
INSERT INTO orders (customer_id, ship_tb, billing, shipping, total_ex_vat, vat_total, total_inc_vat)
  VALUES(1, true, '{"address_id": 1, "contact_name": "Joe Blogs", "addr1": "4524 Mulberry Avenue",
            "city": "LittleRock", "postcode": "72209", "country": "US"}', NULL, 12598, 2520, 15118);
INSERT INTO order_items (order_id, sku, name, qty, unit_price, currency, discount, tax_code, vat)
  VALUES(100002, 'TV-SKU', 'LCD TV System', 1, 12598, 'GBP', null, 'T20', 2520);

-- 100003
INSERT INTO orders (customer_id, ship_tb, billing, shipping, total_ex_vat, vat_total, total_inc_vat)
  VALUES(2, true, '{"address_id": 2, "contact_name": "Sammy Peterson", "addr1": "138 Ermin Street",
            "city": "Wrentham", "postcode": "NR34 9TT", "country": "UK"}', NULL, 9800, 980, 10780);
INSERT INTO order_items (order_id, sku, name, qty, unit_price, currency, discount, tax_code, vat)
  VALUES(100003, 'DRILL-SKU', 'Electric Power Drill', 2, 4900, 'GBP', null, 'T20', 980);

-- 100004
INSERT INTO orders (customer_id, billing, shipping, total_ex_vat, vat_total, total_inc_vat)
  VALUES(3, '{"address_id": 6 , "contact_name": "Faith Bowman", "addr1": "38 Walden Road",
            "city": "Greenburn", "postcode": "DD5 8AU", "country": "UK"}',
            '{"address_id": 3, "contact_name": "Faith Bowman", "addr1": "18 Pier Road",
              "city": "Statham", "postcode": "WA13 3DW", "country": "UK"}', 7849, 1570, 9419);
INSERT INTO order_items (order_id, sku, name, qty, unit_price, currency, discount, tax_code, vat)
  VALUES(100004, 'DESK-SKU', 'Oak Desk', 1, 7849, 'GBP', null, 'T20', 1570);

-- 100005
INSERT INTO orders (customer_id, billing, shipping, total_ex_vat, vat_total, total_inc_vat)
  VALUES(5, '{"address_id": 10, "contact_name": "Bernadette Graham", "addr1": "38 Porana Place",
              "city": "Woolgorong", "postcode": "6630", "country": "AU"}',
            '{"address_id": 9, "contact_name": "Bernadette Graham", "addr1": "89 Cubbine Road",
              "city": "Southburracoppin", "postcode": "6421", "country": "AU"}', 31396, 6279, 37675);
INSERT INTO order_items (order_id, sku, name, qty, unit_price, currency, discount, tax_code, vat)
  VALUES(100005, 'DESK-SKU', 'Oak Desk', 4, 7849, 'GBP', null, 'T20', 6279);


INSERT INTO payments (order_id, typ) VALUES(100001, 'stripe');
INSERT INTO payments (order_id, typ) VALUES(100002, 'paypal');
INSERT INTO payments (order_id, typ) VALUES(100003, 'stripe');
INSERT INTO payments (order_id, typ) VALUES(100003, 'stripe');
INSERT INTO payments (order_id, typ) VALUES(100004, 'paypal');
INSERT INTO payments (order_id, typ) VALUES(100005, 'stripe');

-- products
INSERT INTO products (sku, ean, path, name, content) VALUES('WATER-SKU', 'WATER-EAN', 'water-bottle', 'Water Bottle', '{}');
INSERT INTO products (sku, ean, path, name, content) VALUES('DRILL-SKU', 'DRILL-EAN', 'electric-drill', 'Electric Power Drill', '{}');
INSERT INTO products (sku, ean, path, name, content) VALUES('TV-SKU', 'TV-EAN', 'television-set', 'LCD TV System', '{}');
INSERT INTO products (sku, ean, path, name, content) VALUES('PHONE-SKU', 'PHONE-EAN', 'mobile-phone', 'Mobile Phone Kit', '{}');
INSERT INTO products (sku, ean, path, name, content) VALUES('DESK-SKU', 'DESK-EAN', 'wooden-desk', 'Oak Desk', '{"summary":"Wooden Desk for study","description":"description of desk","specification":"desk spec"}');

-- categories_products
INSERT INTO categories_products (category_id, product_id, path, sku, pri) VALUES (3, 1, 'a/b/e', 'WATER-SKU', 10);
INSERT INTO categories_products (category_id, product_id, path, sku, pri) VALUES (3, 2, 'a/b/e', 'DRILL-SKU', 20);
INSERT INTO categories_products (category_id, product_id, path, sku, pri) VALUES (3, 4, 'a/b/e', 'PHONE-SKU', 30);


-- product pricing tiers
INSERT INTO pricing_tiers(tier_ref, title, description) VALUES ('default', 'Default pricing', '');
INSERT INTO pricing_tiers (tier_ref, title, description)
  VALUES('goldfish', 'Small Wholesale Customer', 'Small company with turn over less than 10k');
INSERT INTO pricing_tiers (tier_ref, title, description)
  VALUES('seabass', 'Medium-sized Customer', 'Medium-sized company with turn over less than 100k');


-- product pricing
-- default
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('default', 'WATER-SKU', 20417);
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('default', 'DRILL-SKU', 395833);
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('default', 'TV-SKU', 2066250);
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('default', 'PHONE-SKU', 241583);
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('default', 'DESK-SKU', 2987083);

-- goldfish
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('goldfish', 'WATER-SKU', 14500);
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('goldfish', 'DRILL-SKU', 152900);
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('goldfish', 'TV-SKU', 1245700);
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('goldfish', 'PHONE-SKU', 145300);
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('goldfish', 'DESK-SKU', 2248200);

-- seabass
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('seabass', 'WATER-SKU', 12900);
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('seabass', 'DRILL-SKU', 122900);
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('seabass', 'TV-SKU', 995700);
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('seabass', 'PHONE-SKU', 125300);
INSERT INTO product_pricing (tier_ref, sku, unit_price) VALUES('seabass', 'DESK-SKU', 1984200);

INSERT INTO shipping_tarrifs (country_code, shipping_code, name, price, tax_code) VALUES ('GB', 'free_delivery', 'Standard Delivery (3-5 working days)', 24583, 'T20');
INSERT INTO shipping_tarrifs (country_code, shipping_code, name, price, tax_code) VALUES ('GB', 'next_day_delivery', 'Next Day Delivery', 41250, 'T20');
INSERT INTO shipping_tarrifs (country_code, shipping_code, name, price, tax_code) VALUES ('GB', 'next_day_pre10', 'Next Day Pre-10:30 Delivery', 124167, 'T20');
INSERT INTO shipping_tarrifs (country_code, shipping_code, name, price, tax_code) VALUES ('GB', 'saturday_delivery', 'Saturday Delivery', 79167, 'T20');
INSERT INTO shipping_tarrifs (country_code, shipping_code, name, price, tax_code) VALUES ('GB', 'sunday_delivery', 'Sunday Delivery', 82500, 'T20');


INSERT INTO shipping_tarrifs (country_code, shipping_code, name, price, tax_code) VALUES ('FR', 'international_standard', 'International Standard (3-5 working days)', 83333, 'T20');
INSERT INTO shipping_tarrifs (country_code, shipping_code, name, price, tax_code) VALUES ('FR', 'international_express', 'International Express (1-2 working days)', 150000, 'T20');

INSERT INTO shipping_tarrifs (country_code, shipping_code, name, price, tax_code) VALUES ('NO', 'international_standard', 'International Standard (3-5 working days)', 120000, 'T0');
INSERT INTO shipping_tarrifs (country_code, shipping_code, name, price, tax_code) VALUES ('NO', 'international_express', 'International Express (1-2 working days)', 220000, 'T0');
