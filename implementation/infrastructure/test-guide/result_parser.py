import os
import re
import json
from datetime import datetime, timezone, timedelta

def parse_time_to_epoch_ms(time_string: str) -> int:
    """
    Parses a time string with a timezone like 'YYYY-MM-DD HH:MM (TZ)'
    and converts it to epoch milliseconds.
    Currently handles WIB (UTC+7).
    """
    # Clean up the string, e.g., "2025-06-16 20:50 (WIB)" -> "2025-06-16 20:50"
    time_part = time_string.split('(')[0].strip()

    # Define the timezone for WIB (Western Indonesia Time), which is UTC+7
    wib_tz = timezone(timedelta(hours=7))

    # Parse the date and time string into a naive datetime object
    naive_dt = datetime.strptime(time_part, "%Y-%m-%d %H:%M")

    # Make the datetime object timezone-aware
    aware_dt = naive_dt.replace(tzinfo=wib_tz)

    # Convert the timezone-aware datetime object to a Unix timestamp (seconds)
    epoch_seconds = aware_dt.timestamp()

    # Convert to milliseconds and return as an integer
    return int(epoch_seconds * 1000)

def find_and_parse_md_files(root_folder: str) -> list[dict]:
    """
    Traverses a folder, finds markdown files starting with 'f',
    parses them, and returns the data as a list of dictionaries.
    """
    all_results = []
    
    # Regex pattern to capture the required fields from the markdown file.
    # Using named capture groups for clarity.
    pattern = re.compile(
        r"Variant: (?P<variant>.*)\n"
        r"Scenario: (?P<scenario>.*)\n"
        r"Flow Control: (?P<flow_control>.*)\n"
        r"Database: (?P<database>.*)\n"
        r"Start Time: (?P<start_time>.*)\n"
        r"End Time: (?P<end_time>.*)",
        re.MULTILINE
    )

    # Walk through the directory tree
    for dirpath, _, filenames in os.walk(root_folder):
        for filename in filenames:
            # Check if the file is a markdown file and starts with 'f'
            if filename.startswith('f') and filename.endswith('.md'):
                file_path = os.path.join(dirpath, filename)
                print(f"Processing file: {file_path}")
                
                try:
                    with open(file_path, 'r', encoding='utf-8') as f:
                        content = f.read()
                    
                    match = pattern.search(content)
                    
                    if not match:
                        print(f"  -> Warning: Could not find matching pattern in {filename}")
                        continue
                        
                    data = match.groupdict()
                    
                    # Extract and structure the data
                    result_entry = {
                        "variant": data["variant"].strip(),
                        "scenario": data["scenario"].strip(),
                        "flow_control": data["flow_control"].strip(),
                        "database": data["database"].strip(),
                        "start_time_str": data["start_time"].strip(),
                        "end_time_str": data["end_time"].strip(),
                        "source_file": filename
                    }
                    
                    # Convert timestamps to epoch milliseconds
                    result_entry["from"] = parse_time_to_epoch_ms(result_entry["start_time_str"])
                    result_entry["to"] = parse_time_to_epoch_ms(result_entry["end_time_str"])

                    all_results.append(result_entry)
                    print(f"  -> Successfully parsed {filename}")

                except Exception as e:
                    print(f"  -> Error processing file {filename}: {e}")

    return all_results

# --- Example Usage ---
if __name__ == "__main__":
    target_folder = "test-notes"
    
    parsed_data = find_and_parse_md_files(target_folder)
    
    with open("tests.json", "w") as w:
        w.write(json.dumps(parsed_data, indent=2))



