# Strava Activities Updater

This project is a console application to update Strava activities' names from russian "Утренний забег" to English "Morning Run".

## How to run

1. Download and install [Go](https://go.dev/dl/).
2. Install `strava-updater`:

    ```sh
    go install github.com/alexandear/strava-updater@v0.0.2
    ```

3. Get Strava access token from https://developers.strava.com/docs/getting-started/.
4. Run:

    ```sh
    strava-updater -accessToken <access_token> -from 2021-01-01 -to 2021-12-31
    ```
