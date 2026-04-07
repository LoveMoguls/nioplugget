-- Insert Geografi and Historia subjects
-- display_order continues from Phase 7: Biologi=1, Samhällskunskap=2, Matematik=3, Kemi=4, Fysik=5
INSERT INTO subjects (name, slug, display_order) VALUES
    ('Geografi', 'geografi', 6),
    ('Historia', 'historia', 7);

-- Geografi topics
INSERT INTO topics (subject_id, name, slug, display_order)
SELECT s.id, t.name, t.slug, t.display_order
FROM subjects s
CROSS JOIN (VALUES
    ('Befolkning och urbanisering', 'befolkning-och-urbanisering', 1),
    ('Klimat och klimatförändringar', 'klimat-och-klimatforandringar', 2),
    ('Naturresurser och hållbarhet', 'naturresurser-och-hallbarhet', 3),
    ('Geopolitik och handel', 'geopolitik-och-handel', 4)
) AS t(name, slug, display_order)
WHERE s.slug = 'geografi';

-- Historia topics
INSERT INTO topics (subject_id, name, slug, display_order)
SELECT s.id, t.name, t.slug, t.display_order
FROM subjects s
CROSS JOIN (VALUES
    ('Industrialismens tid', 'industrialismens-tid', 1),
    ('De två världskrigen', 'de-tva-varldskrigen', 2),
    ('Kalla kriget', 'kalla-kriget', 3),
    ('1900-talets politiska rörelser', '1900-talets-politiska-rorelser', 4)
) AS t(name, slug, display_order)
WHERE s.slug = 'historia';

-- Geografi: Befolkning och urbanisering exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Befolkningsfördelning och migration', 'Förstå push/pull-faktorer och befolkningstäthet.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Urbanisering och globalisering', 'Förstå vad som driver urbanisering och hur globalisering påverkar migration.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Befolkningstillväxt och resurser', 'Resonera kring sambandet mellan befolkningstillväxt, resurser och hållbarhet.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'geografi'
  AND t.slug = 'befolkning-och-urbanisering';

-- Geografi: Klimat och klimatförändringar exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Klimatzoner och klimatfaktorer', 'Förstå jordens klimatzoner och skillnaden mellan klimat och väder.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Global uppvärmning och konsekvenser', 'Förklara sambandet mellan växthusgaser och global uppvärmning.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Klimatåtgärder och ansvarsfördelning', 'Resonera kring länders ansvar för klimatåtgärder och Parisavtalet.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'geografi'
  AND t.slug = 'klimat-och-klimatforandringar';

-- Geografi: Naturresurser och hållbarhet exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Förnybara och icke-förnybara resurser', 'Förstå skillnaden mellan fossila bränslen och förnybara energikällor.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Vattentillgång och energiresurser', 'Förklara varför vattenbrist uppstår och hur energiresurser påverkar geopolitik.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Hållbar resurshushållning', 'Resonera kring hållbar utveckling och SDG-målen.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'geografi'
  AND t.slug = 'naturresurser-och-hallbarhet';

-- Geografi: Geopolitik och handel exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Världshandel och ekonomiska system', 'Förstå vad BNP mäter och vad handelsmönster innebär.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Globala organisationer och beroenden', 'Förklara hur FN och WTO påverkar världshandeln.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Geopolitiska konflikter och Nord-Syd', 'Resonera kring orsaker till ekonomiska klyftor mellan Nord och Syd.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'geografi'
  AND t.slug = 'geopolitik-och-handel';

-- Historia: Industrialismens tid exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Industrialiseringens orsaker och spridning', 'Förstå vad industrialisering innebar och vilka uppfinningar som var centrala.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Sociala rörelser och arbetsvillkor', 'Förklara varför fackföreningar bildades och hur arbetsvillkoren förändrades.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Industrialismens konsekvenser', 'Resonera kring industrialismens långsiktiga konsekvenser för samhälle och miljö.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'historia'
  AND t.slug = 'industrialismens-tid';

-- Historia: De två världskrigen exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Första världskrigets orsaker', 'Förstå de viktigaste orsakerna till första världskriget och allianssystemet.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Andra världskriget och Förintelsen', 'Förklara hur Versaillesfreden bidrog till andra världskrigets utbrott och hur Förintelsen var möjlig.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Världskrigens efterdyningar', 'Resonera kring hur de två världskrigen omformade den internationella ordningen.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'historia'
  AND t.slug = 'de-tva-varldskrigen';

-- Historia: Kalla kriget exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Den bipolära världsordningen', 'Förstå vad kalla kriget innebar och vad bipolär världsordning betyder.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Proxykonflikt och kapprustning', 'Förklara vad proxykonflikt innebär och hur kapprustningen påverkade säkerhetsläget.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Kalla krigets slut och dekolonisation', 'Resonera kring hur kalla kriget påverkade dekolonisationsprocessen.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'historia'
  AND t.slug = 'kalla-kriget';

-- Historia: 1900-talets politiska rörelser exercises
INSERT INTO exercises (topic_id, title, description, difficulty_order, system_prompt)
SELECT t.id, e.title, e.description, e.difficulty_order, e.system_prompt
FROM topics t
JOIN subjects s ON t.subject_id = s.id
CROSS JOIN (VALUES
    ('Fascism och kommunism som ideologier', 'Förstå grunddragen i fascism och kommunism och vad totalitär stat innebär.', 1, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Demokratisering och mänskliga rättigheter', 'Förklara hur FNs deklaration om mänskliga rättigheter kom till och vad som drev demokratiseringsrörelser.', 2, 'SYSTEM_PROMPT_PLACEHOLDER'),
    ('Politiska rörelsers arv', 'Resonera kring hur 1900-talets politiska rörelser formar vår samtid.', 3, 'SYSTEM_PROMPT_PLACEHOLDER')
) AS e(title, description, difficulty_order, system_prompt)
WHERE s.slug = 'historia'
  AND t.slug = '1900-talets-politiska-rorelser';
