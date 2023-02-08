CREATE TABLE comments (
    id VARCHAR ( 64 ) PRIMARY KEY,
    textFr TEXT,
    textEn TEXT,
    publishedAt TIMESTAMP,
    authorId VARCHAR ( 64 ),
    targetId VARCHAR ( 64 ),
    replies VARCHAR ( 64 ) ARRAY NOT NULL DEFAULT array[]::varchar[]
);

INSERT INTO comments VALUES (
    'Comment-kjh784fgevdhhdwhh7563',
    'Bonjour ! je suis un commentaire.',
    'Hi ! Im a comment.',
    TO_TIMESTAMP('1639477064'),
    'User-kjh784fgevdhhdwhh7563',
    'Photo-bdgetr657434hfggrt8374'
);

INSERT INTO comments VALUES (
    'Comment-1234abcd',
    'Je suis une reponse au commentaire',
    'Im a reply!',
    TO_TIMESTAMP('1639477064'),
    'User-5647565dhfbdshs',
    'Comment-kjh784fgevdhhdwhh7563'
);

INSERT INTO comments VALUES (
    'Comment-5678efgh',
    'Je suis une autre reponse !',
    'Im another reply!',
    TO_TIMESTAMP('1639477064'),
    'User-5342hdfgetrfiw789',
    'Comment-kjh784fgevdhhdwhh7563'
);

UPDATE comments SET replies = array_append(replies, 'Comment-1234abcd') WHERE id = 'Comment-kjh784fgevdhhdwhh7563';
UPDATE comments SET replies = array_append(replies, 'Comment-5678efgh') WHERE id = 'Comment-kjh784fgevdhhdwhh7563';