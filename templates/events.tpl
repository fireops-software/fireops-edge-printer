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
# {{.TypEng}}

|   |   |
|---|---|
| **Einsatznummer:** | {{.Num1}} |
| **Alarmiert:** | {{.CreateTime}} |
| **Sirenenprogramm:** | {{.Category}} |
| **Alarmstufe:** | {{.AlarmLev}} |
| **Anrufer:** | {{.CallerName}} |
| **Telefonnummer:** | {{.CallerNumber}} |
| **Ort:** | {{.Location}} |
| **Alarmtext:** | {{.EventAlarmtext}} |
| **Feuerwehren:** | {{range .Destinations }}{{.Name}}, {{end}} |

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

*Powered by FireOPS*

<div style="page-break-after: always;"></div>
{{end}}
