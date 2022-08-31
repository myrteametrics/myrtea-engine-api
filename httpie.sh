
# login
http --session=admin --print=HhBbm :9000/api/v5/engine/security/myself -A bearer -a $(http POST :9000/api/v5/login login=admin password=myrtea | jq -r .token)

# GET /search/last
http --session=admin --print=HhBbm :9000/api/v5/engine/search/last situationid==1 situationinstanceid==1
http --session=admin --print=HhBbm :9000/api/v5/engine/search/last situationid==1
http --session=admin --print=HhBbm :9000/api/v5/engine/search/last situationid==1 mindate==2022-08-01T00:00:00.000+02:00 maxdate==2022-08-18T12:30:00.000+02:00
http --session=admin --print=HhBbm :9000/api/v5/engine/search/last

# GET /search/last/byinterval
http --session=admin --print=HhBbm :9000/api/v5/engine/search/last/byinterval situationid==1 situationinstanceid==3 interval==day mindate==2022-08-01T00:00:00.000+02:00
http --session=admin --print=HhBbm :9000/api/v5/engine/search/last/byinterval situationid==1 situationinstanceid==3 interval==month

# GET /search/last/bycustominterval
http --session=admin --print=HhBbm :9000/api/v5/engine/search/last/bycustominterval situationid==1 situationinstanceid==3 mindate==2022-08-01T00:00:00.000+02:00 interval==12h referencedate==2022-08-31T23:59:59.000+02:00 | jq '.[] | .situations[].dateTime'
http --session=admin --print=HhBbm :9000/api/v5/engine/search/last/bycustominterval situationid==1 situationinstanceid==3 mindate==2022-08-01T00:00:00.000+02:00 interval==12h referencedate==2022-08-31T23:59:59.000+02:00

# GET /issues/*
http --session=admin --print=HhBbm :9000/api/v5/engine/issues/15888
http --session=admin --print=HhBbm :9000/api/v5/engine/issues/15888/history
http --session=admin --print=HhBbm :9000/api/v5/engine/issues/15888/facts_history
http --session=admin --print=HhBbm :9000/api/v5/engine/issues/15912/facts_history

# POST /schedule/trigger
echo -n '{"name":"test","cronexpr":"* * * * *","jobtype":"fact","job":{"factIds":[19, 20]}}' | http --session=admin --print=HhBbm :9000/api/v5/engine/scheduler/trigger

# POST /service/aggregates 
echo -n '[{"factId":19,"situationId":4,"situationInstanceId":94,"time":"2022-08-17T16:07:14.000+02:00","value":{"aggs":{"doc_count":{"value":12345}}}}, {"factId":20,"situationId":4,"situationInstanceId":94,"time":"2022-08-17T16:07:14.000+02:00","value":{"aggs":{"doc_count":{"value":12345}}}}]' | http --session=admin --print=HhBbm :9000/api/v5/service/aggregates 

# POST /service/objects 
echo -n '[{"id": "","index": "","type": "","source": {"key": "8J00625838156", "aggs": {"cp": {"value": "Unspecified"}, "id": {"value": "8J00625838156"}, "risk": {"value": "1"}, "canal": {"value": "Unspecified"}, "poids": {"value": 0.36899998784065247}, "cpdest": {"value": "94170"}, "nb_recla": {"value": 1}, "site_pch": {"value": "009910_VENTE ON LINE"}, "risk_score": {"value": 0.51}, "site_depot": {"value": "580830_CORBIGNY BP"}, "customer_id": {"value": "Boyer_"}, "num_dossier": {"value": "COL-7507870"}, "site_distri": {"value": "Unspecified"}, "code_produit": {"value": "8J"}, "customer_name": {"value": "Boyer"}, "raisonsociale": {"value": "Unspecified"}, "montant_facture": {"value": 8.486999720335007}, "lasttrakingevent": {"value": "DEPGUI"}, "raisonsocialedest": {"value": "Unspecified"}, "assurance_optionnelle": {"value": 0}, "between_pch_and_recla": {"value": 30}}}}]' | http --session=admin --print=HhBbm :9000/api/v5/service/objects fact==tri_annon_taux_count


curl -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NjA4NTE2MDIsImlhdCI6MTY2MDgwODQwMiwiaWQiOiJhODZmYjExYi0wZTAxLTQ2MjItODhkMi0yZWNlZDkwMTZjYjQiLCJpc3MiOiJNeXJ0ZWEgbWV0cmljcyIsIm5iZiI6MTY2MDgwODQwMn0.C96b2HK-avzUV__DmZ0O4C8IlXgRHaqDztlnz6-SIUY' -L --output /dev/null --silent --show-error --write-out 'lookup:        %{time_namelookup}\nconnect:       %{time_connect}\nappconnect:    %{time_appconnect}\npretransfer:   %{time_pretransfer}\nredirect:      %{time_redirect}\nstarttransfer: %{time_starttransfer}\ntotal:         %{time_total}\n' 'http://localhost:9000/api/v5/engine/search/last'
