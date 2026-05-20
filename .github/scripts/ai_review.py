import os
import sys
import json
import urllib.request
import urllib.error

def get_pr_diff():
    """Get the PR diff from GitHub API"""
    repo = os.environ["GITHUB_REPOSITORY"]
    pr_number = os.environ["PR_NUMBER"]
    token = os.environ["GITHUB_TOKEN"]

    url = f"https://api.github.com/repos/{repo}/pulls/{pr_number}/files"
    req = urllib.request.Request(url, headers={
        "Authorization": f"Bearer {token}",
        "Accept": "application/vnd.github.v3+json"
    })

    with urllib.request.urlopen(req) as response:
        files = json.loads(response.read())

    diff_parts = []
    total_lines = 0

    for f in files:
        if total_lines > 500:
            break
        filename = f.get("filename", "")
        patch = f.get("patch", "")
        if patch and any(filename.endswith(ext) for ext in [".go", ".ts", ".tsx", ".py"]):
            diff_parts.append(f"### {filename}\n```\n{patch}\n```")
            total_lines += patch.count("\n")

    return "\n\n".join(diff_parts) if diff_parts else None

def review_with_groq(diff: str) -> str:
    """Send diff to Groq (Llama 3.1) and get security/quality review"""
    import time

    api_key = os.environ["GROQ_API_KEY"]
    url = "https://api.groq.com/openai/v1/chat/completions"

    prompt = f"""You are a security-focused code reviewer for a medical OCR platform handling PHI (Patient Health Information).

Review this PR diff and provide concise feedback focused on:
1. Security vulnerabilities (injection, auth bypass, data exposure)
2. PHI/medical data handling issues
3. Error handling problems
4. Code quality issues

Be specific and reference file names and line numbers.
If the changes look good, say so briefly.
Keep the review under 400 words.

Diff:
{diff}"""

    body = json.dumps({
        "model": "llama-3.1-8b-instant",
        "messages": [{"role": "user", "content": prompt}],
        "max_tokens": 600
    }).encode("utf-8")

    for attempt in range(5):
        try:
            req = urllib.request.Request(url, data=body, headers={
                "Content-Type": "application/json",
                "Authorization": f"Bearer {api_key}"
            })
            with urllib.request.urlopen(req) as response:
                result = json.loads(response.read())
            return result["choices"][0]["message"]["content"]
        except urllib.error.HTTPError as e:
            if e.code == 429:
                wait = 10 * (attempt + 1)
                print(f"Rate limited, waiting {wait}s (attempt {attempt + 1}/5)...")
                time.sleep(wait)
            else:
                raise

    return "⚠️ AI review unavailable. Please review manually."

def post_comment(review: str):
    """Post the review as a PR comment"""
    repo = os.environ["GITHUB_REPOSITORY"]
    pr_number = os.environ["PR_NUMBER"]
    token = os.environ["GITHUB_TOKEN"]

    body = f"## 🤖 AI Security Review (Llama 3.1 via Groq)\n\n{review}\n\n---\n*This review is AI-generated. Human approval is still required before merge.*"

    url = f"https://api.github.com/repos/{repo}/issues/{pr_number}/comments"
    data = json.dumps({"body": body}).encode("utf-8")
    req = urllib.request.Request(url, data=data, headers={
        "Authorization": f"Bearer {token}",
        "Accept": "application/vnd.github.v3+json",
        "Content-Type": "application/json"
    })

    with urllib.request.urlopen(req) as response:
        if response.status == 201:
            print("Review posted successfully")
        else:
            print(f"Failed to post comment: {response.status}")
            sys.exit(1)


def main():
    print("Fetching PR diff...")
    diff = get_pr_diff()

    if not diff:
        print("No reviewable files changed, skipping.")
        return

    print("Sending to Groq for review...")
    review = review_with_groq(diff)

    print("Posting review comment...")
    post_comment(review)

    print("Done.")


if __name__ == "__main__":
    main()
