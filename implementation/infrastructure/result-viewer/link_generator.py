import json

def generate_test_report():
    """
    Reads test data from JSON files and generates a formatted Markdown report.
    """
    try:
        # Load the data from the JSON files
        with open('tests.json', 'r') as f:
            tests = json.load(f)
        with open('agent_dashboards.json', 'r') as f:
            agent_dashboards = json.load(f)
        with open('backend_dashboards.json', 'r') as f:
            backend_dashboards = json.load(f)
    except FileNotFoundError as e:
        print(f"Error: {e}. Make sure tests.json, agent_dashboards.json, and backend_dashboards.json are in the same directory.")
        return
    except json.JSONDecodeError as e:
        print(f"Error decoding JSON: {e}")
        return

    # Define the desired order for test variants
    variant_order = ['stress-2', 'stress-1', 'stress-0', 'sim-1']
    
    # Define database order
    database_order = ['PostgreSQL', 'CitusData', 'YugabyteDB']

    # Group tests by flow control and then by database
    grouped_tests = {
        'nofc': {db: [] for db in database_order},
        'fc': {db: [] for db in database_order}
    }

    for test in tests:
        fc_key = 'nofc' if test.get('flow_control') == 'nofc' else 'fc'
        db = test.get('database')
        # Handle different naming conventions for PostgreSQL
        if db and 'postgres' in db.lower():
            db_key = 'PostgreSQL'
        elif db == 'citusdata':
            db_key = 'CitusData'
        elif db == 'yugabytedb':
             db_key = 'YugabyteDB'
        else:
            db_key = db # Fallback for other databases

        if db_key in grouped_tests[fc_key]:
            grouped_tests[fc_key][db_key].append(test)

    # Start building the Markdown string
    markdown_output = "# Test URLs\n\n"

    # --- Generate "No Flow Control" Section ---
    markdown_output += "## No Flow Control\n\n"
    for db_name in database_order:
        markdown_output += f"### {db_name}\n\n"
        
        # Sort tests according to the specified variant_order
        sorted_tests = sorted(
            grouped_tests['nofc'][db_name],
            key=lambda x: variant_order.index(x['variant']) if x['variant'] in variant_order else len(variant_order)
        )

        if not sorted_tests:
            markdown_output += "No tests found for this configuration.\n\n"
        
        for test in sorted_tests:
            markdown_output += f"#### {test['variant']}\n\n"
            
            # Construct URLs for agent dashboards
            markdown_output += "##### Agent Dashboards\n\n"
            for dashboard in agent_dashboards:
                url = f"{dashboard['url']}?from={test['from']}&to={test['to']}"
                markdown_output += f"- [{dashboard['name']}]({url})\n"
            markdown_output += "\n"

            # Construct URLs for backend dashboards
            markdown_output += "##### Backend Dashboards\n\n"
            for dashboard in backend_dashboards:
                url = f"{dashboard['url']}?from={test['from']}&to={test['to']}"
                markdown_output += f"- [{dashboard['name']}]({url})\n"
            markdown_output += "\n"


    # --- Generate "Flow Control" Section ---
    markdown_output += "## Flow Control\n\n"
    for db_name in database_order:
        markdown_output += f"### {db_name}\n\n"
        
        # Sort tests according to the specified variant_order
        sorted_tests = sorted(
            grouped_tests['fc'][db_name],
            key=lambda x: variant_order.index(x['variant']) if x['variant'] in variant_order else len(variant_order)
        )
        
        if not sorted_tests:
            markdown_output += "No tests found for this configuration.\n\n"

        for test in sorted_tests:
            markdown_output += f"#### {test['variant']}\n\n"
            
            # Construct URLs for agent dashboards
            markdown_output += "##### Agent Dashboards\n\n"
            for dashboard in agent_dashboards:
                url = f"{dashboard['url']}?from={test['from']}&to={test['to']}"
                markdown_output += f"- [{dashboard['name']}]({url})\n"
            markdown_output += "\n"

            # Construct URLs for backend dashboards
            markdown_output += "##### Backend Dashboards\n\n"
            for dashboard in backend_dashboards:
                url = f"{dashboard['url']}?from={test['from']}&to={test['to']}"
                markdown_output += f"- [{dashboard['name']}]({url})\n"
            markdown_output += "\n"


    # Write the output to a markdown file
    with open('result_urls.md', 'w') as f:
        f.write(markdown_output)

    print("Successfully generated test_urls.md")


if __name__ == "__main__":
    generate_test_report()
