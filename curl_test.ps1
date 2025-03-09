# List of verses
$verses = @(
    "1Peter 1:6"
    "1John 4:13-17"
    "1Corinthians 16:13"
    "love"
    "hardship"
    "Jesus"
    "the"
    "Midian"
    "beloved's"
    "Judges 1:1"
    "love"
    "hate"
    "and"
    "there"
)

while ($true) {
    $verses | ForEach-Object -ThrottleLimit 10 -Parallel {
        $verse = $_
        $encoded_verse = $verse -replace ' ', '%20'
        $url = "http://localhost/api?searchMode=stringsearch?searchString=$encoded_verse%20?caseSensitive=true"
        Write-Output "Fetching data for verse: $encoded_verse"
        Invoke-RestMethod -Method Get -Uri $url
        Write-Output "`n"
    }
}
