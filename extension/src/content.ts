const isGitHubTab = window.location.hostname === "github.com";

if (isGitHubTab) {
  console.log("Meridian content script active on GitHub");
}
