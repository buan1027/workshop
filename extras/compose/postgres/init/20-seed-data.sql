INSERT INTO gebrauchtwagen.gebrauchtwagen (
    fin,
    marke,
    modell,
    baujahr,
    erstzulassung,
    kilometerstand,
    kraftstoffart,
    fahrzeugklasse,
    ausstattung,
    schadenfrei,
    version
) VALUES
    ('WVWZZZ1JZXW000001', 'VW', 'Golf', 2020, '2020-06-15', 42000, 'BENZIN', 'KOMPAKTKLASSE', '{"klimaautomatik": true}', true, 1),
    ('WBA8E31090K000002', 'BMW', '320d', 2019, '2019-03-20', 83000, 'DIESEL', 'MITTELKLASSE', '{"navigation": true}', false, 1),
    ('WAUZZZF3XKN000003', 'Audi', 'A4 Avant', 2021, '2021-09-01', 31000, 'HYBRID', 'KOMBI', '{"anhaengerkupplung": true}', true, 1)
ON CONFLICT (fin) DO NOTHING;

INSERT INTO gebrauchtwagen.standort (plz, ort, gebrauchtwagen_id)
SELECT '76131', 'Karlsruhe', id
FROM gebrauchtwagen.gebrauchtwagen
WHERE fin = 'WVWZZZ1JZXW000001'
ON CONFLICT (gebrauchtwagen_id) DO NOTHING;

INSERT INTO gebrauchtwagen.hauptuntersuchung (
    pruefdatum,
    gueltig_bis,
    prueforganisation,
    status,
    gebrauchtwagen_id
)
SELECT '2025-06-01', '2027-06-01', 'TUEV', 'BESTANDEN', id
FROM gebrauchtwagen.gebrauchtwagen
WHERE fin = 'WVWZZZ1JZXW000001'
ON CONFLICT (gebrauchtwagen_id) DO NOTHING;

INSERT INTO gebrauchtwagen.schaden (
    bezeichnung,
    beschreibung,
    feststellungsdatum,
    gebrauchtwagen_id
)
SELECT 'Kratzer', 'Kleiner Lackkratzer an der hinteren Tuer', '2024-11-10', id
FROM gebrauchtwagen.gebrauchtwagen
WHERE fin = 'WBA8E31090K000002';
