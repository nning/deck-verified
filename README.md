# deck-verified

Keeps track of Steam Deck Verifications. On first run, it reports all games with
their respective Steam Deck Verification status. On subsequent runs, the tool
will report newly tested and updated games.

## Commands

### update

     $ ./deck-verified update
     New      DEATH STRANDING  Verified
     Updated  DEATHLOOP        Verified

     Total: 202, New: 1, Updated: 1

### search

    $ ./deck-verified search tomb raider
    Tomb Raider: Legend                            Playable  2019-12-12 01:18:46 +0100 CET
    Rise of the Tomb Raider™                       Playable  2019-07-02 12:33:44 +0200 CEST
    Tomb Raider                                    Playable  2020-05-15 16:05:30 +0200 CEST
    Shadow of the Tomb Raider: Definitive Edition  Playable  2019-12-19 18:33:08 +0100 CET
    Tomb Raider IV: The Last Revelation            Playable  2019-12-12 01:20:53 +0100 CET

### list

List can be filtered by `unsupported`, `playable`, and `verified`.

    $ ./deck-verified list verified | head -3
    DEATH STRANDING                                     Verified  2020-12-15 09:31:46 +0100 CET
    Bayonetta                                           Verified  2020-11-27 20:34:00 +0100 CET
    Business Tour - Board Game with Online Multiplayer  Verified  2021-11-23 10:16:04 +0100 CET

