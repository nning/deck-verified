<table>
    <tr>
        <td>Name</td>
        <td><strong>{{ .Name }}</strong></td>
    </tr>
    <tr>
        <td>Status</td>
        <td><strong>{{ .Status }}</strong></td>
    </tr>
    <tr>
        <td>App ID</td>
        <td>{{ .AppID }}</td>
    </tr>
    <tr>
        <td>Last Updated (SteamDB)</td>
        <td>{{ .LastUpdatedSteamDB }}</td>
    </tr>
    <tr>
        <td>Last Updated (Here)</td>
        <td>{{ .LastUpdatedHere }}</td>
    </tr>
    <tr>
        <td>First Seen</td>
        <td>{{ .FirstSeen }}</td>
    </tr>
    <tr>
        <td>Links</td>
        <td>
            <a href="https://steamdb.info/app/{{ .AppID }}/info/">SteamDB</a><br>
            <a href="https://www.protondb.com/app/{{ .AppID }}">ProtonDB</a><br>
            <a href="https://store.steampowered.com/app/{{ .AppID }}">Steam Store</a>
        </td>
    </tr>
</table>