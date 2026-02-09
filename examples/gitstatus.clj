#!/usr/bin/env bb

(ns git-status
  (:require [babashka.fs :as fs]
            [babashka.process :as p]
            [cheshire.core :as json]))

(def config
  {:plugin {:id "git-status"
            :title "Git Status"
            :version "0.1.0"}
   :projects-dir (str (or (System/getenv "AUTOMATE_ME_PROJECTS_DIR")
                          (str (fs/home) fs/file-separator "projects")))} )

(defn list-projects [dir]
  (->> (fs/list-dir dir)
       (filter fs/directory?)
       (map #(fs/file-name %))
       sort
       vec))

(defn manifest []
  (let [choices (if (fs/exists? (:projects-dir config))
                  (list-projects (:projects-dir config))
                  [])]
    {:schemaVersion 1
     :plugin (:plugin config)
     :tasks [{:name "gst"
              :title "Git status"
              :group "Git"
              :description "Show git status for a project"
              :inputs [{:name "project"
                        :type "enum"
                        :required true
                        :prompt "Project"
                        :choices choices}]}]}))

(defn read-json-stdin []
  (let [raw (slurp *in*)]
    (if (seq (clojure.string/trim raw))
      (json/parse-string raw true)
      {})))

(defn run-command [command args]
  (let [project (:project args)]
    (when-not (seq project)
      (binding [*out* *err*]
        (println "missing required input: project"))
      (System/exit 1))
    (let [path (str (:projects-dir config) fs/file-separator project)]
      (when-not (fs/directory? path)
        (binding [*out* *err*]
          (println (str "project not found: " path)))
        (System/exit 1))
      (p/shell {:dir path} command))))

(defn usage []
  (binding [*out* *err*]
    (println "usage: plugin describe|run <task>"))
  (System/exit 1))

(defn -main [& args]
  (let [cmd (first args)]
    (case cmd
      "describe" (println (json/generate-string (manifest)))
      "run" (let [task (second args)
                  payload (read-json-stdin)
                  task-args (get payload :args {})]
              (case task
                "gst" (run-command "git status" task-args)
                (do (binding [*out* *err*]
                      (println (str "unknown task: " task)))
                    (System/exit 1))))
      (usage))))

(apply -main *command-line-args*)
