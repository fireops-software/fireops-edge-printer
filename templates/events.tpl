<style>
  table {
    width: 100%;
    border: 1px solid black;
    border-collapse: collapse;
  }

  tr {
    height: 32px;
    border: 1px solid black;
    border-collapse: collapse;
  }

  td {
    border: 1px solid black;
    border-collapse: collapse;
  }

</style>

{{range .}}
# Einsatz von FireOps

|   |   |
|---|---|
| **Einsatznummer:** | {{.Num1}} |
| **Zeitstempel:** | {{.CreateTime}} |
| **Kategorie:** | {{.Category}} |
| **Art:** | {{.TypEng}} {{.SubEng}} |
| **Alarmstufe:** | {{.AlarmLev}} |
| **Anrufer:** | {{.CallerName}} |
| **Telefonnummer:** | {{.CallerNumber}} |
| **Ort:** | {{.Location}} |
| **Ortsinfo:** | {{.LocationInfo}} |
| **Alarmtext:** | {{.EventAlarmtext}} |
| **Destinations:** | {{range .Destinations }}{{.Name}}, {{end}} |

## Mannschaft

| Vorname | Nachname | Einheit |
|---|---|---|
| | |
| | |
| | |
| | |
| | |
| | |
| | |
| | |
| | |

<div style="page-break-after: always;"></div>
{{end}}
