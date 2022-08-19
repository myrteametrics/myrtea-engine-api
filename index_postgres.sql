\d situation_history_v5
\d fact_history_v5
\d situation_fact_history_v5

create index idx_fact_history_v5_combo on fact_history_v5 (fact_id, ts desc) include (id);
create index idx_situation_fact_history_v5_situation_history_id on situation_fact_history_v5 (situation_history_id);
create index idx_situation_history_v5_combo on situation_history_v5 (situation_id, situation_instance_id, ts desc) include (id);

explain analyze SELECT fh.*, f.name FROM fact_history_v5 fh INNER JOIN fact_definition_v1 f on fh.fact_id = f.id WHERE fh.fact_id = 1 ORDER BY fh.ts desc LIMIT 1;
explain analyze SELECT fh.*, f.name FROM fact_history_v5 fh INNER JOIN fact_definition_v1 f on fh.fact_id = f.id WHERE fh.id IN (1,2,3);

explain analyze SELECT * FROM situation_fact_history_v5 WHERE situation_history_id IN (1, 2, 3, 4, 5);

explain analyze SELECT id FROM situation_history_v5 ;
explain analyze SELECT id FROM situation_history_v5 where situation_id = 1;
explain analyze SELECT id FROM situation_history_v5 where situation_id = 1 and situation_instance_id = 4;
explain analyze SELECT id FROM situation_history_v5 where situation_id = 1 and situation_instance_id = 4 and ts > '2022-08-01' and ts < '2022-08-29';
explain analyze SELECT distinct on (situation_id, situation_instance_id) id FROM situation_history_v5 ORDER BY situation_id, situation_instance_id, ts desc;
explain analyze SELECT distinct on (situation_id, situation_instance_id) id FROM situation_history_v5 where situation_id = 1 ORDER BY situation_id, situation_instance_id, ts desc;
explain analyze SELECT distinct on (situation_id, situation_instance_id) id FROM situation_history_v5 where situation_id = 1 and situation_instance_id = 4 ORDER BY situation_id, situation_instance_id, ts desc;
explain analyze SELECT distinct on (situation_id, situation_instance_id) id FROM situation_history_v5 where situation_id = 1 and situation_instance_id = 4 and ts > '2022-08-01' and ts < '2022-08-29' ORDER BY situation_id, situation_instance_id, ts desc;
explain analyze SELECT distinct on (situation_id, situation_instance_id, date_trunc('day', ts)) id FROM situation_history_v5 ORDER BY situation_id, situation_instance_id, date_trunc('day', ts) desc;
explain analyze SELECT distinct on (situation_id, situation_instance_id, CAST('2022-08-18T17:38:45+02:00' AS TIMESTAMPTZ) + INTERVAL '1 second' * 172800 * FLOOR(DATE_PART('epoch', ts- '2022-08-18T17:38:45+02:00')/172800)) id FROM situation_history_v5 ORDER BY situation_id, situation_instance_id, CAST('2022-08-18T17:38:45+02:00' AS TIMESTAMPTZ) + INTERVAL '1 second' * 172800 * FLOOR(DATE_PART('epoch', ts- '2022-08-18T17:38:45+02:00')/172800) desc;
explain analyze SELECT sh.*, s.name, si.name FROM situation_definition_v1 s LEFT JOIN situation_template_instances_v1 si on s.id = si.situation_id INNER JOIN situation_history_v5 sh on (s.id = sh.situation_id and (sh.situation_instance_id = si.id OR sh.situation_instance_id = 0)) WHERE sh.id in (SELECT distinct on (situation_id, situation_instance_id) id FROM situation_history_v5 ORDER BY situation_id, situation_instance_id, ts desc);



drop index if exists idx_situation_history_v5_ts; 
drop index if exists idx_situation_history_v5_situation_id; 
drop index if exists idx_situation_history_v5_situation_instance_id; 
drop index if exists idx_situation_history_v5_combo;

create index idx_situation_history_v5_ts on situation_history_v5 (ts);
create index idx_situation_history_v5_situation_id on situation_history_v5 (situation_id);
create index idx_situation_history_v5_situation_instance_id on situation_history_v5 (situation_instance_id);
create index idx_situation_history_v5_combo on situation_history_v5 (situation_id, situation_instance_id, ts desc nulls last);


explain analyze select distinct on (s.situation_id, s.situation_instance_id) id from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc;


drop index if exists idx_situation_history_v5_id; 
drop index if exists idx_fact_history_v5_id; 
drop index if exists idx_situation_fact_history_v5_combo; 
drop index if exists idx_situation_fact_history_v5_situation_history_id;
drop index if exists idx_situation_fact_history_v5_fact_history_id;
create index idx_situation_history_v5_id on situation_history_v5 (id);
create index idx_fact_history_v5_id on fact_history_v5 (id);

create index idx_fact_history_v5_fact_id on fact_history_v5 (fact_id);
create index idx_situation_fact_history_v5_situation_history_id on situation_fact_history_v5 (situation_history_id);
create index idx_situation_fact_history_v5_fact_history_id on situation_fact_history_v5 (fact_history_id);
create index idx_situation_fact_history_v5_combo on situation_fact_history_v5 (situation_history_id, fact_history_id);


explain analyze select * 
from situation_history_v5 s
inner join situation_fact_history_v5 sf on s.id = sf.situation_history_id
inner join fact_history_v5 f on sf.fact_history_id = f.id
where s.situation_id = 3;

explain analyze select * from situation_history_v5 where situation_id = 3 and ts > '2022-08-01';
explain analyze select distinct on (s.situation_id) s.situation_id, ts from situation_history_v5 s order by s.situation_id, s.ts desc;
explain analyze select distinct on (s.situation_id, s.situation_instance_id) s.situation_id, s.situation_instance_id, ts from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc;
explain analyze select distinct on (s.situation_id, s.situation_instance_id) * from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc;
explain analyze select distinct on (s.situation_id, s.situation_instance_id) * from situation_history_v5 s where ts > now() - '14 days'::interval order by s.situation_id, s.situation_instance_id, s.ts desc;
explain analyze select distinct on (s.situation_id, s.situation_instance_id) id from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc;
explain analyze select * from situation_history_v5 where id in (select distinct on (s.situation_id, s.situation_instance_id) id from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc);



explain select * from situation_history_v5 s where s.id in (select distinct on (s.situation_id, s.situation_instance_id) id from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc);
explain select * from situation_history_v5 s inner join situation_fact_history_v5 sf on s.id = sf.situation_history_id inner join fact_history_v5 f on sf.fact_history_id = f.id ;
explain select * from situation_history_v5 s inner join situation_fact_history_v5 sf on s.id = sf.situation_history_id inner join fact_history_v5 f on sf.fact_history_id = f.id where s.situation_id = 3;
explain select f.* from fact_history_v5 f inner join situation_fact_history_v5 sf on f.id = sf.fact_history_id where sf.situation_history_id in (select distinct on (s.situation_id, s.situation_instance_id) id from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc);

drop index if exists idx_situation_history_v5_ts; 
drop index if exists idx_situation_history_v5_situation_id; 
drop index if exists idx_situation_history_v5_combo;
explain analyze select distinct on (s.situation_id, s.situation_instance_id) * from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc;

create index idx_situation_history_v5_ts on situation_history_v5 (ts);
explain analyze select distinct on (s.situation_id, s.situation_instance_id) * from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc;

create index idx_situation_history_v5_situation_id on situation_history_v5 (situation_id);
explain analyze select distinct on (s.situation_id, s.situation_instance_id) * from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc;

drop index if exists idx_situation_history_v5_combo;
create index idx_situation_history_v5_combo on situation_history_v5 (situation_id, situation_instance_id, ts desc);
explain analyze select distinct on (s.situation_id, s.situation_instance_id) s.situation_id, s.situation_instance_id, ts from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc;

drop index if exists idx_situation_history_v5_combo;
create index idx_situation_history_v5_combo on situation_history_v5 (situation_id, situation_instance_id, ts desc) include (id);
explain analyze select distinct on (s.situation_id, s.situation_instance_id) id from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc;



drop index if exists idx_fact_history_v5_ts; 
drop index if exists idx_fact_history_v5_fact_id; 
drop index if exists idx_fact_history_v5_combo; 
create index idx_fact_history_v5_ts on fact_history_v5 (ts);
create index idx_fact_history_v5_fact_id on fact_history_v5 (fact_id);
create index idx_fact_history_v5_combo on fact_history_v5 (fact_id, ts desc) include (id);
explain analyze select distinct on (f.fact_id) id from fact_history_v5 f order by f.fact_id, f.ts desc;




-- get last value (all situation / all instance)
explain analyze select * from situation_history_v5 where id in (select distinct on (s.situation_id, s.situation_instance_id) id from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc);
-- with situation_definition
explain analyze select sh.* from situation_history_v5 sh inner join situation_definition_v1 sd on sh.situation_id = sd.id where sh.id in (select distinct on (s.situation_id, s.situation_instance_id) id from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc);
explain analyze select * from fact_history_v5 where id in (select fact_history_id from situation_fact_history_v5 sf where sf.situation_history_id in (select distinct on (s.situation_id, s.situation_instance_id) id from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc));
explain analyze select * from fact_history_v5 where id in (select fact_history_id from situation_fact_history_v5 sf where sf.situation_history_id in (211057,651462,2488209,2488312,2488383,2487499,2487458,2487519,651498,2488500,2488216,2488576,2488450,651318,2488541,2488246,2488294,2487538,2488358,2488411,2488419,651081,2488345,651518,2488410,2488293,2488478,2488396,651092,2487449,2488228,2488258,2488322,2488955,2488230,2488476,2487389,2488401,2488259,2488937,2488962,2488219,2488386,651251,2488210,2488203,651101,651388,2488496,2487573,2488972,2488943,2488373,2488939,651120,2488916,2487450,651200,2488283,2488507,2488328,2488301,2488928,2488565,2484257,2488913,651185,2488940,2488233,2487456,2488460,2488435,2487474,2488924,2488907,2488430,651100,2488511,2488976,2487459,2488477,2488398,2488929,2487505,2488442,2488296,2487508,2488282,2488919,2488549,2488321,2488506,2487479,651188,2487390,2488264,651175,2488509,2488215,2487486,2487509,2488288,2487529,2488484,2488968,2487457,2488920,2488526,2488926,651394,2488518,2488287,1796921,2488289,2488914,2488953,2488956,2487567,2488302,2488250,2488959,2488529,2488584,2487580,2488448,2487537,2487556,2488342,2487566,2488319,2488375,2488327,2488933,1481322,651541));
explain analyze select * from fact_history_v5 f inner join situation_fact_history_v5 sf on f.fact_id = sf.fact_history_id where sf.situation_history_id in (211057,651462,2488209,2488312,2488383,2487499,2487458,2487519,651498,2488500,2488216,2488576,2488450,651318,2488541,2488246,2488294,2487538,2488358,2488411,2488419,651081,2488345,651518,2488410,2488293,2488478,2488396,651092,2487449,2488228,2488258,2488322,2488955,2488230,2488476,2487389,2488401,2488259,2488937,2488962,2488219,2488386,651251,2488210,2488203,651101,651388,2488496,2487573,2488972,2488943,2488373,2488939,651120,2488916,2487450,651200,2488283,2488507,2488328,2488301,2488928,2488565,2484257,2488913,651185,2488940,2488233,2487456,2488460,2488435,2487474,2488924,2488907,2488430,651100,2488511,2488976,2487459,2488477,2488398,2488929,2487505,2488442,2488296,2487508,2488282,2488919,2488549,2488321,2488506,2487479,651188,2487390,2488264,651175,2488509,2488215,2487486,2487509,2488288,2487529,2488484,2488968,2487457,2488920,2488526,2488926,651394,2488518,2488287,1796921,2488289,2488914,2488953,2488956,2487567,2488302,2488250,2488959,2488529,2488584,2487580,2488448,2487537,2487556,2488342,2487566,2488319,2488375,2488327,2488933,1481322,651541);
explain analyze select * from fact_history_v5 fh inner join fact_definition_v1 on fh.fact_id = f.id where id in (732463,732688,732640,732435,732277,2560474,2560578,2560398,2561226,2560844,2560300,2560816,2561147,2560592,2560718,2560323,2560598,2560105,2560985,2560794,2560430,2560665,2559799,2559684,2560347,2560193,2561211,2560956,2560391,2560787,2560137,2560329,2559938,2560667,2560705,2560068,2560335,2560973,2559668,2561480,2560603,2560403,2559981,2560712,2560215,2560779,2561220,2559798,2560534,2559782,2560533,2560061,2560880,2561167,2560963,2560341,2559649,2560462,2561487,2560856,2560075,2561213,2559797,2560560,2560786,2559676,2560062,2560894,2560475,2561090,2560404,2560143,2561030,2560991,2559982,2561079,2560633,2560012,2559609,2561295,2560639,2560112,2560056,2561413,2560850,2560506,2560822,2559675,2560992,2560774,3856015,4455138,5998209,5997908,4997385,6000076,5998868,6000187,6000042,5999791,5999120,5999608,5999416,6000154,6000157,5999365,5999667,5999479,5999916,5999308,5999172,6000080,5999368,5999045,5999949,6000078,5999912,5999788,5999044,5999954,6000072,5999613,5999617,5999792,5999369,6000190,6000039,5999619,5999474,5999472,5999914,5999544,5999917,5999665,5999737,5999736,5999837,5999175,5999953,5999952,5999839,6000120,5999414,6000188,5999622,5999311,6000153,5999616,6000081,5999997,5999115,5999838,5999541,5999671,5999843,5999842,6000124,5999043,6000077,6000075,6000118,5999787,6000038,6000082,6000083,5999908,5999911,5999367,5999418,5999612,5999480,5999614,5999904,6000079,5999793,6000155,6000189,5999789,6000123,5999611,5999473,5999546,6000156,6000119,5999478,5999468,5999739,5999119,5999620,5999844,5999177,5999475,5999742,5999540,5999610,6000036,5999118,5999176,6000084,5999545,5999229,5999674,5999669,5999989,5999909,5999178,6000040,5999615,5999840,5998867,5999362,5999735,5999790,5999228,5999744,5999245,6000201,6000135,6000073,6000099,6000269,5999857,5999687,5999690,6000138,6000229,6000018,5999811,6000055,5999688,5999489,5999488,5999609,6000014,5999686,6000054,5998984,6000137,5999929,6000015,6000256,5999430,5999558,5999560,6000202,6000231,6000056,5999491,6000270,6000140,6000139,5999864,5999808,5999563,6000016,5999930,6000017,6000255,5999247,5999809,6000271,6000230,6000144,5999052,6000254,6000203,5999432,5999756,6000058,6000141,5999692,5999860,5999862,5999562,5999861,5999967,5999053,6000136,6000288,5999494,5999965,6000274,5999565,5999495,5999055,5999056,5999434,6000103,6000311,6000178,6000257,6000143,5999566,5999813,6000179,6000309,5999863,5999935,4997494,5999976,6000150,4997334,6000263,5999879,6000068,6000338,6000067,6000240,4997553,6000293,4997328,6000239,4997737,5999503,5999881,6000339,6000287,4997166,6000343,5999977,5999875,4997855,6000151,5999876,4997226,6000212,6000307,6000297,4997448,6000242,6000116,4997850,5999877,4997620,6000211,6000341,4997613,6000210,5999975,4997614,6000340,6000070,4997004,6000264,4997618,6000069,5999878,4997851,6000345,6000344,4997854,6000308,6000186,4997335,5999880,6000330,6000025,4997801,6000071,4997107,6000329,6000213,6000152,4997444,6000241
-- get last value (situation x / all instance)
explain analyze select * from situation_history_v5 where id in (select distinct on (s.situation_id, s.situation_instance_id) id from situation_history_v5 s where s.situation_id = 3 order by s.situation_id, s.situation_instance_id, s.ts desc);
-- get last value (situation x / instance x)
explain analyze select * from situation_history_v5 where id in (select distinct on (s.situation_id, s.situation_instance_id) id from situation_history_v5 s where s.situation_id = 3 and s.situation_instance_id = 61 order by s.situation_id, s.situation_instance_id, s.ts desc);


-- get full history by interval + reference date (all situation / all instance)
explain analyze select distinct on (s.situation_id, s.situation_instance_id, CAST('2022-08-01' AS TIMESTAMP) + INTERVAL '1 second' * 86400 * FLOOR(DATE_PART('epoch', ts- '2022-08-01')/86400)) id 
from situation_history_v5 s order by s.situation_id, s.situation_instance_id, CAST('2022-08-01' AS TIMESTAMP) + INTERVAL '1 second' * 86400 * FLOOR(DATE_PART('epoch', ts- '2022-08-01')/86400) desc;

-- get full history by standard interval (all situation / all instance)
explain analyze select distinct on (s.situation_id, s.situation_instance_id, date_trunc('hour', ts)) id 
from situation_history_v5 s order by s.situation_id, s.situation_instance_id, date_trunc('hour', ts) desc;


-- useless as fields are already primary key
-- drop index if exists idx_situation_definition_v1_id; 
-- create index idx_situation_definition_v1_id on situation_definition_v1 (id);
-- 
-- drop index if exists idx_situation_template_instance_v1_id; 
-- create index idx_situation_template_instance_v1_id on situation_template_instances_v1 (id);


explain analyze select sh.*
from situation_definition_v1 sd
left join situation_template_instances_v1 si on sd.id = si.situation_id 
inner join situation_history_v5 sh on (
    sd.id = sh.situation_id 
    and (sh.situation_instance_id = si.id OR sh.situation_instance_id = 0)
)
where sh.id in (select distinct on (s.situation_id, s.situation_instance_id) id from situation_history_v5 s order by s.situation_id, s.situation_instance_id, s.ts desc);












// func TestQuery2(t *testing.T) {
    // 	t.Fail()
    // 	db := tests.DBClient(t)
    
    // 	shq := HistorySituationsQuerier{conn: db}
    // 	options := GetHistorySituationsOptions{SituationID: 3, SituationInstanceID: -1, FromTS: time.Date(2022, time.July, 1, 0, 0, 0, 0, time.UTC)}
    // 	// historySituationsSQL, args, err := shq.GetHistorySituationsLast(options).ToSql()
    // 	historySituationsSQL, args, err := shq.GetHistorySituationsByStandardInterval(options, "day").ToSql()
    // 	// historySituationsSQL, args, err := shq.GetHistorySituationsByCustomInterval(options, time.Date(2022, time.July, 1, 0, 0, 0, 0, time.UTC), 24 * time.Hour).ToSql()
    // 	if err != nil {
    // 		t.Log(err)
    // 		return
    // 	}
    // 	t.Log(historySituationsSQL, args)
    
    // 	historySituationsQuery := shq.GetHistorySituationsDetails(historySituationsSQL, args)
    // 	historySituationsSQL, args, err2 := historySituationsQuery.ToSql()
    // 	if err2 != nil {
    // 		t.Log(err2)
    // 		return
    // 	}
    // 	t.Log(historySituationsSQL, args)
    
    // 	t.Log("start query")
    // 	ctx, ctxCancel := context.WithTimeout(context.Background(), 2*time.Second)
    // 	defer ctxCancel()
    // 	rows, err := historySituationsQuery.QueryContext(ctx)
    // 	if err != nil {
    // 		t.Log(err)
    // 		return
    // 	}
    // 	defer rows.Close()
    
    // 	t.Log("query done")
    // 	historySituationsIds := make([]int64, 0)
    // 	for rows.Next() {
    // 		item := HistorySituationsV4{}
    // 		err := rows.Scan(
    // 			&item.ID,
    // 			&item.SituationID,
    // 			&item.SituationInstanceID,
    // 			&item.Ts,
    // 			&item.Parameters,
    // 			&item.ExpressionDacts,
    // 			&item.Metadatas,
    // 		)
    // 		if err != nil {
    // 			t.Error(err)
    // 		}
    
    // 		historySituationsIds = append(historySituationsIds, item.ID)
    // 	}
    
    // 	fhq := HistoryFactsQuerier{conn: db}
    // 	rows, err = fhq.GetHistoryFactsBuilder(historySituationsIds).Query()
    // 	if err != nil {
    // 		t.Log(err)
    // 		return
    // 	}
    // 	defer rows.Close()
    
    // 	historyFactsIds := make([]int64, 0)
    // 	for rows.Next() {
    // 		item := HistorySituationFactsV4{}
    // 		err := rows.Scan(
    // 			&item.HistorySituationID,
    // 			&item.HistoryFactID,
    // 		)
    // 		if err != nil {
    // 			t.Error(err)
    // 		}
    
    // 		historyFactsIds = append(historyFactsIds, item.HistoryFactID)
    // 	}
    
    // 	fhq.GetHistoryFactsLast(historyFactsIds)
    // }
    
    func TestQuery(t *testing.T) {
        t.Fail()
        db := tests.DBClient(t)
        psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)
        // sqlLastSituation, _, err2 := psql.
        // 	Select("id").
        // 	From("situation_history_v5").
        // 	// Where(sq.Eq{"situation_id": nil}).          // Optionnal SituationID
        // 	// Where(sq.Eq{"situation_instance_id": nil}). // Optionnal SituationInstanceID
        // 	// Where(sq.GtOrEq{"ts": ""}).                 // Optionnal Minimum Date
        // 	// Where(sq.Lt{"ts": ""}).                     // Optionnal Maximum Date
        // 	Options("distinct on (situation_id, situation_instance_id)").
        // 	OrderBy("situation_id", "situation_instance_id", "ts desc").
        // 	ToSql()
        // if err2 != nil {
        // 	t.Log(err2)
        // }
        // t.Log(sqlLastSituation)
    
        // sqlSituationByStandardInterval, _, err3 := psql.
        // 	Select("id").
        // 	From("situation_history_v5").
        // 	// Where(sq.Eq{"situation_id": nil}).          // Optionnal SituationID
        // 	// Where(sq.Eq{"situation_instance_id": nil}). // Optionnal SituationInstanceID
        // 	// Where(sq.GtOrEq{"ts": ""}).                 // Optionnal Minimum Date
        // 	// Where(sq.Lt{"ts": ""}).                     // Optionnal Maximum Date
        // 	Options("distinct on (situation_id, situation_instance_id, date_trunc('hour', ts))").
        // 	OrderBy("situation_id", "situation_instance_id", "date_trunc('hour', ts)").
        // 	ToSql()
        // if err3 != nil {
        // 	t.Log(err3)
        // }
        // t.Log(sqlSituationByStandardInterval)
    
        // sqlSituationByCustomInterval, _, err3 := psql.
        // 	Select("id").
        // 	From("situation_history_v5").
        // 	// Where(sq.Eq{"situation_id": nil}).          // Optionnal SituationID
        // 	// Where(sq.Eq{"situation_instance_id": nil}). // Optionnal SituationInstanceID
        // 	// Where(sq.GtOrEq{"ts": ""}).                 // Optionnal Minimum Date
        // 	// Where(sq.Lt{"ts": ""}).                     // Optionnal Maximum Date
        // 	Options("distinct on (situation_id, situation_instance_id, CAST('2022-08-01' AS TIMESTAMP) + INTERVAL '1 second' * 86400 * FLOOR(DATE_PART('epoch', ts- '2022-08-01')/86400))").
        // 	OrderBy("situation_id", "situation_instance_id", "CAST('2022-08-01' AS TIMESTAMP) + INTERVAL '1 second' * 86400 * FLOOR(DATE_PART('epoch', ts- '2022-08-01')/86400)").
        // 	ToSql()
        // if err3 != nil {
        // 	t.Log(err3)
        // }
        // t.Log(sqlSituationByCustomInterval)
    
        // sqlHistorySituationsDetails, _, err := psql.
        // 	Select("sh.*").
        // 	From("situation_definition_v1 s").
        // 	LeftJoin("situation_template_instances_v1 si on s.id = si.situation_id").
        // 	InnerJoin("situation_history_v5 sh on (s.id = sh.situation_id and (sh.situation_instance_id = si.id OR sh.situation_instance_id = 0))").
        // 	Where("sh.id in (" + sqlLastSituation + ")").
        // 	// Where("sh.id in (" + sqlSituationByStandardInterval+ ")").
        // 	// Where("sh.id in (" + sqlSituationByCustomInterval + ")").
        // 	ToSql()
        // if err != nil {
        // 	t.Log(err)
        // }
        // t.Log(sqlHistorySituationsDetails)
    
        ids := []interface{}{2488288, 2487529, 2488484, 2488968, 2487457, 2488920, 2488526}
        sqlHistoryFactsDetailsSub, vars, _ := psql.Select("fact_history_id").From("situation_fact_history_v5").Where(sq.Eq{"situation_history_id": ids}).ToSql()
        t.Log(sqlHistoryFactsDetailsSub)
        t.Log(vars)
        // select fact_history_id from situation_fact_history_v5 sf where sf.situation_history_id in (1, 2)
        sqlHistoryFactsDetails, _, err := psql.
            Select("*").
            From("fact_history_v5").
            Where("id in ("+sqlHistoryFactsDetailsSub+")", ids...).
            // Where("id in (select fact_history_id from situation_fact_history_v5 sf where sf.situation_history_id in (" + sqlLastSituation + "))").
            // Where("id in (select fact_history_id from situation_fact_history_v5 sf where sf.situation_history_id in (" + sqlSituationByStandardInterval+ "))").
            // Where("id in (select fact_history_id from situation_fact_history_v5 sf where sf.situation_history_id in (" + sqlSituationByCustomInterval + "))").
            ToSql()
        if err != nil {
            t.Log(err)
        }
        t.Log(sqlHistoryFactsDetails)
    
        rows, err := psql.Select("*").
            From("fact_history_v5").
            Where("id in ("+sqlHistoryFactsDetailsSub+")", ids...).
            Query()
        if err != nil {
            t.Error(err)
            return
        }
        defer rows.Close()
    
        for rows.Next() {
            item := HistoryFactsV4{}
            err := rows.Scan(
                &item.ID,
                &item.FactID,
                &item.SituationID,
                &item.SituationInstanceID,
                &item.Ts,
                &item.Result,
            )
            if err != nil {
                t.Error(err)
            }
            t.Log(item)
        }
    }
    