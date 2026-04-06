-- Insert Kemi and Fysik subjects
-- display_order continues from existing: Biologi=1, Samhällskunskap=2, Matematik=3
INSERT INTO subjects (name, slug, display_order) VALUES
    ('Kemi', 'kemi', 4),
    ('Fysik', 'fysik', 5);

-- Kemi topics
INSERT INTO topics (subject_id, name, slug, display_order)
SELECT s.id, t.name, t.slug, t.display_order
FROM subjects s
CROSS JOIN (VALUES
    ('Materiens uppbyggnad', 'materiens-uppbyggnad', 1),
    ('Kemiska reaktioner', 'kemiska-reaktioner', 2),
    ('Kemikalier i vardagen', 'kemikalier-i-vardagen', 3),
    ('Kemi och hållbar utveckling', 'kemi-och-hallbar-utveckling', 4)
) AS t(name, slug, display_order)
WHERE s.slug = 'kemi';

-- Fysik topics
INSERT INTO topics (subject_id, name, slug, display_order)
SELECT s.id, t.name, t.slug, t.display_order
FROM subjects s
CROSS JOIN (VALUES
    ('Krafter och rörelse', 'krafter-och-rorelse', 1),
    ('Elektricitet och magnetism', 'elektricitet-och-magnetism', 2),
    ('Energi och energiomvandlingar', 'energi-och-energiomvandlingar', 3),
    ('Astronomi och universum', 'astronomi-och-universum', 4)
) AS t(name, slug, display_order)
WHERE s.slug = 'fysik';

-- Kemi: Materiens uppbyggnad exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Atomer och grundämnen', 'Förstå atomens uppbyggnad och periodiska systemet.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Molekyler och kemiska bindningar', 'Förstå hur atomer binds samman till molekyler och föreningar.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Ämnen och egenskaper', 'Resonera kring hur ämnets struktur påverkar dess egenskaper.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'kemi'
  AND t.slug = 'materiens-uppbyggnad';

-- Kemi: Kemiska reaktioner exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Syror och baser', 'Förstå skillnaden mellan syror och baser och pH-skalan.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Förbränning och oxidation', 'Förstå förbränningsreaktioner och oxidationsprocesser.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Kemiska reaktioners villkor', 'Resonera kring faktorer som påverkar reaktionshastighet och jämvikt.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'kemi'
  AND t.slug = 'kemiska-reaktioner';

-- Kemi: Kemikalier i vardagen exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Lösningar och blandningar', 'Förstå hur ämnen löser sig och bildar blandningar.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Kemikaliesäkerhet', 'Känna till farosymboler och säker hantering av kemikalier.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Kemikalier i samhället', 'Resonera kring kemikaliers roll och konsekvenser i samhället.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'kemi'
  AND t.slug = 'kemikalier-i-vardagen';

-- Kemi: Kemi och hållbar utveckling exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Fossila bränslen och klimat', 'Förstå fossila bränslens påverkan på klimat och miljö.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Alternativa energikällor', 'Jämföra förnybara och icke-förnybara energikällor ur kemiperspektiv.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Hållbar kemianvändning', 'Resonera kring hur kemi kan bidra till en hållbar utveckling.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'kemi'
  AND t.slug = 'kemi-och-hallbar-utveckling';

-- Fysik: Krafter och rörelse exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Newtons rörelselagar', 'Förstå Newtons tre rörelselagar och deras tillämpningar.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Rörelse och hastighet', 'Beräkna och resonera kring rörelse, hastighet och acceleration.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Krafter i vardagen', 'Analysera krafter i vardagliga situationer med fysikaliska modeller.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'fysik'
  AND t.slug = 'krafter-och-rorelse';

-- Fysik: Elektricitet och magnetism exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Elektrisk krets', 'Förstå hur en enkel elektrisk krets fungerar med komponenter.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Ström, spänning och resistans', 'Tillämpa Ohms lag och räkna på ström, spänning och resistans.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Elektromagnetism och induktion', 'Resonera kring sambandet mellan elektricitet och magnetism.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'fysik'
  AND t.slug = 'elektricitet-och-magnetism';

-- Fysik: Energi och energiomvandlingar exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Energiformer och omvandlingar', 'Förstå olika energiformer och hur energi omvandlas.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Energikällor och verkningsgrad', 'Jämföra energikällors verkningsgrad och miljöpåverkan.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Energi och hållbarhet', 'Resonera kring energianvändning ur ett hållbarhetsperspektiv.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'fysik'
  AND t.slug = 'energi-och-energiomvandlingar';

-- Fysik: Astronomi och universum exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Solsystemet', 'Förstå solsystemets uppbyggnad och planeternas rörelser.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Stjärnor och galaxer', 'Beskriva stjärnors livscykel och galaxers struktur.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Universums ursprung', 'Resonera kring Big Bang-teorin och universums uppkomst och utveckling.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'fysik'
  AND t.slug = 'astronomi-och-universum';
