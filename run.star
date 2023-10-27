def run(plan):
  result = plan.run_sh(run="mkdir -p /src && echo kurtosis > /src/tech.txt", store=["/src/tech.txt"])
  file_artifacts = result.files_artifacts
  result2 = plan.run_sh(run="cat /temp/tech.txt", files={"/temp": file_artifacts[0]})
  plan.verify(result2.output, "==", "kurtosis\n")
  result2 = plan.run_sh(run="cat /temp/tech.txt | tr -d '\n'", files={"/temp": file_artifacts[0]})
  plan.verify(result2.output, "==", "kurtosis") # this should pass but fails