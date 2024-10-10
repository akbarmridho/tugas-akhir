import os
import random
from datetime import datetime, timedelta
from github import Github

# GitHub authentication
github_token = os.environ.get("GITHUB_TOKEN")
g = Github(github_token)

# Repository details
repo_name = "akbarmridho/tugas-akhir"
repo = g.get_repo(repo_name)


copypasta_options = [
    {
        "title": "Thesis Deadline Approaches: Panic Intensifies",
        "content": "Your thesis deadline is getting closer faster than you can say \"I'll start tomorrow.\" Remember, future you is going to be really upset with present you if you don't start writing NOW. Don't let future you down - they're counting on you!",
    },
    {
        "title": "Coffee: The Real Thesis Fuel",
        "content": "Legend has it that if you arrange your coffee stains just right, they'll form a perfect outline for your thesis. Time to test this theory! ☕📚 #ThesisLife #CaffeinatedScholar",
    },
    {
        "title": "Citation Needed: A Thesis Story",
        "content": "In the beginning, there was an idea. Then came research, followed by more research, and even more research. Now you're drowning in a sea of papers, and you can't remember where you read that one crucial piece of information. May the gods of proper citation have mercy on your soul.",
    },
    {
        "title": "Who Let Me Do a Thesis Anyway?",
        "content": "Remember: your committee allowed you to do this thesis because they believe in you. Whether that was a grave error in judgment remains to be seen. Prove them right (or at least don't prove them horrendously wrong)!",
    },
    {
        "title": "You Can Do This: A Reminder",
        "content": "Your thesis doesn't have to change the world. It just has to be done. You've come this far - don't give up now. Channel your inner academic warrior and conquer that chapter!",
    },
    {
        "title": "Netflix and No Chill: The Thesis Edition",
        "content": """Dear Netflix, YouTube, and social media,
It's not you, it's me. I need some time alone with my thesis. We can catch up after my defense.
XOXO,
A Struggling Grad Student""",
    },
    {
        "title": "Post-Thesis Life: Myth or Reality?",
        "content": "Rumor has it there's a world outside your research bubble. Some say you'll be able to read for pleasure again, have a social life, and remember what sunlight feels like. Keep going - that world is waiting for you!",
    },
    {
        "title": "Gotta Write It All: Thesis Version",
        "content": """Your thesis used PROCRASTINATE. It's super effective!
You used PANIC WRITING. It's not very effective...
Your advisor used DEADLINE REMINDER. Critical hit!
Will you use ACTUALLY WRITE or CHANGE TOPIC? Choose wisely, trainer!""",
    },
    {
        "title": "Lose Weight with This One Weird Trick: Write a Thesis!",
        "content": "Side effects may include: caffeine addiction, spontaneous crying, ability to subsist on ramen, and intimate knowledge of your library's operating hours. Consult your academic advisor before beginning this regimen. Results may vary.",
    },
    {
        "title": "I-It's not like I want you to finish your thesis or anything... Baka!",
        "content": "Listen up, Akbar-kun! Don't get the wrong idea, but... your thesis won't write itself, you know? Not that I care or anything! But maybe you should stop procrastinating and actually do some work. Hmph!",
    },
    {
        "title": "Your lack of progress is... disappointing.",
        "content": "...Akbar. Your thesis. It requires attention. Neglecting it is... unwise. I suggest you rectify this situation. Immediately. Failure is not an option. Do you understand? Good. Now, return to your work. I'll be watching.",
    },
    {
        "title": "Who You Gonna Call? Thesis Busters!",
        "content": " Warning: Your thesis may be haunted by the ghosts of unfinished chapters, spectral citations, and the lurking presence of imposter syndrome. Do not attempt to exorcise these spirits alone. Call the Thesis Busters hotline now! (Disclaimer: Service may or may not be provided by sleep-deprived grad students in lab coats.)",
    },
]


# Check for recent commits
def check_recent_commits():
    one_week_ago = datetime.now() - timedelta(days=7)
    commits = repo.get_commits(since=one_week_ago)
    return commits.totalCount > 0


# Create an issue
def create_issue():
    sample = random.sample(copypasta_options, 1)[0]
    body = f"""
    @akbarmridho

    {sample['content']}

    Best regards,
    Your Friendly Reminder
    """
    repo.create_issue(title=sample["title"], body=body)


# Main function
def main():
    create_issue()
    # if not check_recent_commits():
    #     create_issue()
    #     print("Reminder issue created.")
    # else:
    #     print("Recent commits found. No action needed.")


if __name__ == "__main__":
    main()
