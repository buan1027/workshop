param(
    [string]$PlantUMLJar = "C:\Zimmermann\plantuml\plantuml.jar"
)

$ErrorActionPreference = "Stop"

New-Item -ItemType Directory -Force -Path docs\html | Out-Null
New-Item -ItemType Directory -Force -Path docs\diagramme\generated | Out-Null

if (Test-Path $PlantUMLJar) {
    java -jar $PlantUMLJar -tsvg -o ..\generated docs\diagramme\src\*.puml
} else {
    Write-Warning "PlantUML-JAR nicht gefunden: $PlantUMLJar. Diagramme werden nicht gerendert."
}

if (Get-Command asciidoctor -ErrorAction SilentlyContinue) {
    asciidoctor -D docs\html docs\projekthandbuch.adoc
} else {
    Write-Warning "asciidoctor ist nicht installiert. Projekthandbuch bleibt als AsciiDoc verfuegbar."
}
