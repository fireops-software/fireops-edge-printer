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
| **Einsatzort:** | {{.Location}}{{with .LocationInfo}} <br> {{.}}{{end}}{{with .LocationInvolved}} <br> Betroffen: {{.}}{{end}} |
| **Alarmtext:** | {{.EventAlarmtext}} |
| **Feuerwehren:** | {{range .Destinations }}{{.Name}}, {{end}} |

![](data:image/png;base64,{{.MapImgBase64}})

*Powered by FireOPS*

<div style="page-break-after: always;"></div>
{{end}}
