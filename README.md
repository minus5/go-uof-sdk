# go-uof-sdk
Betradar Unified Odds Feed Go SDK

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/minus5/go-uof-sdk?status.svg)](https://godoc.org/github.com/minus5/go-uof-sdk) 
[![Go Report Card](https://goreportcard.com/badge/github.com/minus5/go-uof-sdk)](https://goreportcard.com/report/github.com/minus5/go-uof-sdk)
[![Build Status](https://travis-ci.com/minus5/go-uof-sdk.svg)](https://travis-ci.com/minus5/go-uof-sdk)
[![Coverage Status](https://coveralls.io/repos/github/minus5/go-uof-sdk/badge.svg?branch=master)](https://coveralls.io/github/minus5/go-uof-sdk?branch=master)

  

### Why

From the Betradar [docs](https://docs.betradar.com/display/BD/UOF+-+SDK): 

SDK benefits over protocol/API
 * The SDK hides the separation between messages and API-lookups. The client system just receives the message objects where all information is prefilled by the SDK caches, and in some cases looked up in the background when the client system requests some more rarely requested information.
 * The SDK takes care of translations transparently. The client system just needs to define what languages it needs.
 * The SDK takes care of dynamic text markets and outright markets automatically, which requires some extra logic and lookups for someone not using the SDK.
 * The SDK handles initial connect and state, as well as recovery in case of a temporary disconnect. This needs to be handled manually by someone not using the SDK.


### How

SDK se spaja na dva Betradar endpointa. Prvi je AMQP message queue, drugi je rest api. U kodu im odgovaraju paketi queue i api.   Prvim kanalom, queue, dobijamo numericke, brzo promjenjive podatke. Najznacajniji predstavnik su promjene tecajeva. Drugi kanal je vezan us staticke, opisne podatke, koji tipicno ovise o jeziku u kojem ih prezentiramo. Npr. nazivi timova, igraca, razrada. Podaci u queue kanalu, kako su po prirodi numericki, nisu vezani uz neki jezik.

Izlaz je serijal [poruka](https://github.com/minus5/go-uof-sdk/blob/00bca10f295f31c1581826411412ffb1913edf80/message.go#L42). Imamo tri grupe tipova poruka:
 * event messages; poruke vezane uz neki dogadjaj:
   * odds change
   * fixture change; npr vrijeme pocetka
   * bet cancel
   * bet settlement
   * bet stop
   * rollback bet settlement
   * rollback bet stop
 * api message; poruke koje su dogovor na neki api poziv:
   * fixture - detalji dogadjaja
   * markets - opsi razrada
   * player  - opis igraca
 * system messages
   * alive
   * snapshot complete
   * connection status
   * producer state change

Za client aplikaciju nuzno je da razumije i da ne odgovarajuci nacin reagira na svaki tip poruke.

Treci tip poruka, system messages, je nuzno razumjeti za ostvarivanje pouzdane komunikacije.  
Alive i snapshot complete su low level poruke. Od klijenta se ne ocekuje da reagira ne njih. Slijedeca dva tipa connection status i producer state change su puno znacajniji. Alive i snapshot pustamo van da bi klijent imao kompletnu sliku komunikacije, da bi mogao logirati i kasnije eventualno istrazivati probleme.

Connection status ima dva stanja up ili down. Up je ako smo uspjesno spojeni na Betradar AMQP message queue, down ako nismo. Kada izgubimo konekciju SDK ce emitirati down poruku, kada ju ponovo upsostavimo saljemo up. Slicno kao i prethodne dvije poruke na ovu takodjer nije nuzno reagirati ima informativno znacenje.

Sto je [producer](https://docs.betradar.com/display/BD/UOF+-+Producers)?  
Sve event message types su vezane uz nekog producera. Producer je npr; live, prematch, svaki od virtuala. Produceri neovisno proizvode i salju poruke vezano uz dogadjaje koje pokrivaju.

Bitan tip sistemske poruke je producer state. U njoj dobijamo stanje *svih* producera na svaku promjenu bilo kojeg od njih. Svaki producer moze bit u jednom od tri stanja:
 * down
 * active
 * in recovery

Inicijalno svi produceri startaju u in recovery stanju. Kada recovery procedura zavrsi producer prelazi u active stanje. Bilo koji gubitak veze prebacuje sve producere u down stanje. Uspostava veze stavlja producere u in recovery stanje i pokrece recovery proceduru za sve. 

#### Recovery procedura

Ogovornost je klijenta (korisnika SDK) da za *svakog producera* pamti zadnji timestamp poruke koja je uspjesno obradjena. Na pokretanju aplikacije SDK se inicijalizira tim timestamp-om za svakog producera kojeg klijent prati.  
SDK ce na startu pokrenuti recovery proceduru za svako producera. Kijent ce na samom startu dobiti producer change poruku u kojoj su svi produceri u in recovery stanju. Kako zavrsi recovery za nekog od producera klijent ce dobiti producers change poruku u kojoj je promjenjeno stanje tog producera.  
Od klijent aplikacije se ocekuje da reagira na producers change poruke na odgovarajuci nacin. Npr. da zaustavi kladnjenje na evente producera kada je on u stanju koje nije active duze od 20 sekundi.  
Uputno je razumjeti SDK <-> Betradar komunikaciju tijekom recovery-ja: [referenca](https://docs.betradar.com/display/BD/UOF+-+Recovery+using+API)


### Primjer

koristenja SDK:
```Go
	err := sdk.Run(ctx,
		sdk.Credentials(bookmakerID, token),		
		sdk.Languages(uof.Languages("en,de,hr")),        
		sdk.Recovery(pc),
		sdk.Consumer(myConsumer),
	)
	if err != nil {
		log.Fatal(err)
	}
```
Credentials je nuzna postavka da bi se SDK znao spojiti na queue i api.  
Languages defirnira zeljene jezike za statefull poruke.  
Recovery omogucuje da za svaki producer postavimo zadnji uspjesno obradjeni timestamp od kojeg zelimo pokrenuti recovery proceduru.  
Consumer je mjesto gdje klijent konzumira SDK poruke.  
Primjer consumera:

```Go
    myConsumer(in <-chan *uof.Message) error) {
		for msg := range in {
            // handle msg
            // on fatal return error
		}
        // SDK is disconnected
		return nil
	}
```

Consumer cita sve poruke iz in kanala i obradi ih.  
Clean exit se radi na nacin da klijent zatvori ctx koji je poslan u _Run_. Nakon toga ce SDK prekinuti queue konekciju, zavrsiti obradu svih do tada primljenih pourka i zatvoriti _in_ kanale svih  consumera. Nakon sto svi consumeri zavrse zavrsava i _Run_ poziv.  
Fatal, unclean, izlazak je moguc na nacin da handler vrati error prije nego je konzumirao sve poruke iz in kanala.


### Examples

For sample staging client see cmd/client.  
For sample replay see cmd/replay.  

These examples require two env variables to be set:  
```shell
export UOF_BOOKMAKER_ID=...  
export UOF_TOKEN=...  
```

### Code linting

This project uses [golangci-lint](https://github.com/golangci/golangci-lint) linter with config in [.golangci.yml](https://github.com/minus5/go-uof-sdk/blob/master/.golangci.yml).

If you are using vscode add to your settings.json:
```JSON
"go.lintTool":"golangci-lint",
"go.lintFlags": [
  "--fast"
]
```

or run it in terminal:
```shell
golangci-lint run ./...
```

The main difference with old golint is that we have disabled annoying warnings about not having a comment on exported (method|function|type|const) and package.

### Notes

* Staging environment weekend downtime:

The integration environment is available 24/5, Monday to Friday.  
During the weekend there will be some planned 2 hour disconnections at fixed times:  
Saturday: 14:00 - 16:00 UTC and 20:00 - 22:00 UTC  
Sunday: 00:00 - 02:00 UTC and 13:00 - 15:00 UTC  


