-- Events table
CREATE TABLE events (
    event_id TEXT PRIMARY KEY,
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE slugs (
    slug TEXT PRIMARY KEY CHECK (slug ~ '^[a-z0-9-]+$'),
    event_id TEXT NOT NULL REFERENCES events(event_id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE questions (
    question_id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL REFERENCES events(event_id) ON DELETE CASCADE,
    big_text TEXT NOT NULL,
    small_text TEXT NOT NULL,
    image_filename TEXT NOT NULL,
    choice_a TEXT NOT NULL,
    choice_b TEXT NOT NULL
);

CREATE INDEX idx_questions_event_id ON questions(event_id);

CREATE TABLE sessions (
    session_id TEXT PRIMARY KEY,
    name TEXT,
    email TEXT,
    mobile TEXT
);

CREATE TABLE responses (
    question_id TEXT NOT NULL REFERENCES questions(question_id) ON DELETE CASCADE,
    session_id TEXT NOT NULL REFERENCES sessions(session_id) ON DELETE CASCADE,
    slug TEXT NOT NULL REFERENCES slugs(slug) ON DELETE CASCADE,
    choice CHAR(1) NOT NULL CHECK (choice IN ('a', 'b')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (question_id,session_id)
);

-- Insert initial event data
INSERT INTO events (event_id, description) VALUES 
    ('event_39aJ1km3pr9v1yQYX5gS88e3CUM', 'Total Kombat 3');

-- Insert multiple slugs for testing different traffic sources
INSERT INTO slugs (slug, event_id) VALUES 
    ('tk03', 'event_39aJ1km3pr9v1yQYX5gS88e3CUM'),
    ('tk03-stadium', 'event_39aJ1km3pr9v1yQYX5gS88e3CUM'),
    ('tk03-web', 'event_39aJ1km3pr9v1yQYX5gS88e3CUM'),
    ('tk03-test', 'event_39aJ1km3pr9v1yQYX5gS88e3CUM');

-- Insert questions (KSUIDs are chronologically ordered)
INSERT INTO questions (question_id, event_id, big_text, small_text, image_filename, choice_a, choice_b) VALUES
    ('question_39aJ1eE9ihQ3hH9kmOfKdCSueFP', 'event_39aJ1km3pr9v1yQYX5gS88e3CUM', 
     'Joe vs Bahaa', 
     'England vs Spain. Kickboxing vs Taekwondo. 2 current World Champions. Joe Brooks challenges Bahaa Kabil for the first Male Total Kombat World Title in history.', 
     'matchup-joe-vs-bahaa.png', 
     'Joe Brooks', 
     'Bahaa Kabil'),
     
    ('question_39aJ1lU5dd430R5uChL3QSIg9r8', 'event_39aJ1km3pr9v1yQYX5gS88e3CUM', 
     'Tony vs Sid', 
     'The rematch everyone has been waiting for. The debate will be settled the right way - Inside the Oval. Tony "The Axe Man" Stephenson promises another viral knockout. Sid Williams says otherwise. Who takes it?', 
     'matchup-tony-vs-sid.png', 
     'Tony Stephenson', 
     'Sid Williams'),
     
    ('question_39aJ1lcWNS9J0c9WODFGxAgzHXR', 'event_39aJ1km3pr9v1yQYX5gS88e3CUM', 
     'Dee vs Monique', 
     'A clash of styles. Two aggressive fighters. Will Muay Thai reign or will MMA leave victorious? Begley returns to the oval once again with her eyes on victory. Ettiene promises to shine in her debut. Who takes the W?', 
     'matchup-mon-vs-dee.png', 
     'Dee Begley', 
     'Monique Ettiene'),
     
    ('question_39aJ1qegWKFjX7HTNfaXPv5YQ1n', 'event_39aJ1km3pr9v1yQYX5gS88e3CUM', 
     'Luke vs Nathan', 
     'Two families embedded in combat sports. Which family name will leave with the Kombat honour? We have Karate vs Kickboxing. Essex vs Bolton. This one is going to be tasty.', 
     'matchup-luke-vs-nathan.png', 
     'Luke', 
     'Nathan'),
     
    ('question_39aJ1mHGXklu13dBft4kg4zsppy', 'event_39aJ1km3pr9v1yQYX5gS88e3CUM', 
     'Luke Daly vs Tim Bos', 
     'Irish pride vs Italian flare. Kickboxing vs Taekwondo. This fight is going to be electric. Proud Irishman Daly comes in as the underdog. Tim Bos has the weight of Taekwondo on his shoulders. Which style will be victorious?', 
     'matchup-luke-vs-tim.png', 
     'Luke Daly', 
     'Tim Bos'),
     
    ('question_39aJ1pvTXJhj6NKxZcUFfJm9jr2', 'event_39aJ1km3pr9v1yQYX5gS88e3CUM', 
     'Marcus vs Tai', 
     'Liverpool''s notorious Marcus Lewis makes his Total Kombat debut. Will Tai upset the home crowd? Marcus Lewis is ready for war. Tai Gordon has the oval experience. This fight is going to be a banger!', 
     'matchup-mar-vs-tai.png', 
     'Marcus Lewis', 
     'Tai Gordon');
