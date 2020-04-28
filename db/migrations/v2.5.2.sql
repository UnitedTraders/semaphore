ALTER TABLE task add `arguments` text null;
ALTER TABLE `project__template` ADD slug varchar(64) not null default 'someslug' after id ;
UPDATE `project__template` set slug = CONCAT('slug-', CAST(id AS CHAR));
ALTER TABLE `project__template` ADD UNIQUE (slug);
