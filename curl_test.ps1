# List of verses
$verses = @(
    "1John 4:7-12"
    "Lamentations 3:22-23"
    "2Corinthians 4:17"
    "Matthew 6:27"
    "Romans 12:9-14"
    "Romans 12:15-19"
    "Romans 8:38-39"
    "Psalm 46:5"
    "John 16:33"
    "Matthew 6:34"
    "Romans 12:21"
    "Isaiah 30:18"
    "1John 4:18-19"
    "John 8:11"
    "Titus 3:4-5"
    "Romans 12:2"
    "Matthew 17:20"
    "1Peter 4:8"
    "Isaiah 25:8"
    "Romans 8:2"
    "Romans 8:3-5"
    "Romans 8:6-11"
    "John 1:7-19"
    "Matthew 11:28-30"
    "1Peter 1:6"
    "1John 4:13-17"
    "1Corinthians 16:13"
    "love"
    "hardship"
    "Jesus"
    "the"
    "Midian"
    "beloved's"
)

foreach ($verse in $verses) {
    # Replace spaces with + for URL encoding
    $encoded_verse = $verse -replace ' ', '+'
    $url = "http://localhost/api/$encoded_verse"
    Write-Output "Fetching data for verse: $encoded_verse"
    Invoke-RestMethod -Method Get -Uri $url
    Write-Output "`n"
}