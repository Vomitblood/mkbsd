const URL: &str = "https://storage.googleapis.com/panels-api/data/20240916/media-1a-i-p~s";

#[derive(serde::Deserialize)]
struct Data {
    data: std::collections::HashMap<String, SubProperty>,
}

#[derive(serde::Deserialize)]
struct SubProperty {
    dhd: Option<String>,
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("Fetching data from: {URL}");
    let response = fetch_data(URL)?;
    let json_data: Data = parse_data(&response)?;

    let image_count = count_images(&json_data);
    println!("Total images to download: {}", image_count);

    let download_dir = "downloads";
    create_directory(download_dir)?;

    // create the progress bar
    let bar = indicatif::ProgressBar::new(image_count as u64);
    bar.set_style(
        indicatif::ProgressStyle::with_template(
            "{spinner} [{pos}/{len}] {wide_bar} {percent_precise}% ({eta_precise})",
        )
        .unwrap()
        .progress_chars("##-"),
    );

    let client = reqwest::blocking::Client::new();
    let mut file_index = 1;
    let mut success_count = 0;

    for subproperty in json_data.data.values() {
        if let Some(image_url) = &subproperty.dhd {
            match download_image(&client, image_url, download_dir, file_index) {
                Ok(_) => {
                    bar.inc(1);
                    file_index += 1;
                    success_count += 1;
                }
                Err(e) => eprintln!("Error downloading image: {}", e),
            }
        }
        bar.tick();
    }

    bar.finish();
    println!(
        "{}/{} images downloaded successfully",
        success_count, image_count
    );
    Ok(())
}

fn fetch_data(url: &str) -> Result<String, reqwest::Error> {
    reqwest::blocking::get(url)?.text()
}

fn parse_data(response: &str) -> Result<Data, serde_json::Error> {
    serde_json::from_str(response)
}

fn count_images(json_data: &Data) -> usize {
    json_data
        .data
        .values()
        .filter(|subprop| subprop.dhd.is_some())
        .count()
}

fn create_directory(dir: &str) -> std::io::Result<()> {
    std::fs::create_dir_all(dir)?;
    Ok(())
}

fn download_image(
    client: &reqwest::blocking::Client,
    image_url: &str,
    download_dir: &str,
    file_index: usize,
) -> Result<(), Box<dyn std::error::Error>> {
    let response = client.get(image_url).send()?;
    if !response.status().is_success() {
        return Err(format!("Failed to download image: {}", response.status()).into());
    }

    let image_url_without_params = image_url.split('?').next().unwrap_or(image_url);
    let file_extension = std::path::Path::new(image_url_without_params)
        .extension()
        .and_then(|ext| ext.to_str())
        .map(|ext| format!(".{}", ext))
        .unwrap_or_default();

    let file_path = format!("{}/{}{}", download_dir, file_index, file_extension);

    let mut file = std::fs::File::create(&file_path)?;
    std::io::Write::write_all(&mut file, &response.bytes()?)?;

    Ok(())
}
