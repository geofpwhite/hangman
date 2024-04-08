from selenium import webdriver
import time

browser = webdriver.Chrome()
browser2 = webdriver.Chrome()
browser3 = webdriver.Chrome()
browser.get("https://geofpwhite.github.io/hangman")
x = browser.find_element("id", "new-game")
x.click()
print(x)
browser2.get("https://geofpwhite.github.io/hangman")
browser3.get("https://geofpwhite.github.io/hangman")
browser2.refresh()
time.sleep(5)
x2 = browser2.find_element("id", "join-game-0")
x3 = browser3.find_element("id", "join-game-0")
x2.click()
x3.click()
print(x2)

# ex, ex2, ex3 = browser.find_element("id", "exit-game"), browser2.find_element(
#     "id", "exit-game"), browser3.find_element("id", "exit-game")
#
# ex3.click()
# ex2.click()
# ex.click()
