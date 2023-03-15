-- =============================================================================
-- Diagram Name: test
-- Created on: 14.03.2023 12:05:11
-- Diagram Version: 
-- =============================================================================

CREATE TABLE "news" (
	"newsId" SERIAL NOT NULL,
	"title" varchar(255) NOT NULL,
	"preview" varchar(255),
	"content" text,
	"categoryId" int4 NOT NULL,
	"tagIds" int4[],
	"createdAt" timestamp with time zone NOT NULL DEFAULT NOW(),
	"publishedAt" timestamp with time zone,
	"statusId" int4 NOT NULL,
	PRIMARY KEY("newsId")
);

COMMENT ON COLUMN "news"."newsId" IS 'id новости';

COMMENT ON COLUMN "news"."title" IS 'Заголовок новости';

COMMENT ON COLUMN "news"."preview" IS 'Ссылка на фото-превью новости';

COMMENT ON COLUMN "news"."content" IS 'Контент новости';

CREATE TABLE "statuses" (
	"statusId" SERIAL NOT NULL,
	PRIMARY KEY("statusId")
);

CREATE TABLE "categories" (
	"categoryId" SERIAL NOT NULL,
	"title" varchar(255) NOT NULL,
	"orderNumber" int4 NOT NULL,
	"statusId" int4 NOT NULL,
	PRIMARY KEY("categoryId")
);

CREATE TABLE "tags" (
	"tagId" SERIAL NOT NULL,
	"title" varchar(255) NOT NULL,
	"statusId" int4 NOT NULL,
	PRIMARY KEY("tagId")
);

COMMENT ON COLUMN "tags"."tagId" IS 'id тега';

COMMENT ON COLUMN "tags"."title" IS 'Текст тега';


ALTER TABLE "news" ADD CONSTRAINT "Ref_news_to_statuses" FOREIGN KEY ("statusId")
	REFERENCES "statuses"("statusId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;

ALTER TABLE "news" ADD CONSTRAINT "Ref_news_to_categories" FOREIGN KEY ("categoryId")
	REFERENCES "categories"("categoryId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;

ALTER TABLE "categories" ADD CONSTRAINT "Ref_categories_to_statuses" FOREIGN KEY ("statusId")
	REFERENCES "statuses"("statusId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;

ALTER TABLE "tags" ADD CONSTRAINT "Ref_tags_to_statuses" FOREIGN KEY ("statusId")
	REFERENCES "statuses"("statusId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;


