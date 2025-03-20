INSERT INTO "statuses" ( "statusId", "title", "alias" ) VALUES ( 1, 'Опубликован', 'enabled' );
INSERT INTO "statuses" ( "statusId", "title", "alias" ) VALUES ( 2, 'Не опубликован', 'disabled' );
INSERT INTO "statuses" ( "statusId", "title", "alias" ) VALUES ( 3, 'Удален', 'deleted' );

-- password is 12345
INSERT INTO "users" ( "login", "password", "statusId" ) VALUES ( 'admin', '$2y$14$4IpqlaJ2Rvfgs.wb8f6lPODVLb/Ygl6zw1ZCUKz5CuT6WB6CV44AG', 1 );

INSERT INTO "vfsFolders" ("parentFolderId", title, "isFavorite", "createdAt", "statusId") VALUES (null, 'root', false, now(), 1);
